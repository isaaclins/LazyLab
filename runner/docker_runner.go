package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/isaaclins/LazyLab/config"
	"github.com/isaaclins/LazyLab/dockerargs"
)

const (
	defaultImage = "homebrew/brew:latest"
)

var dockerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)

// Run starts a container per config, prepares it, opens interactive shell, and cleans up
func Run(ctx context.Context, cfg config.RuntimeConfig) error {
	containerName := cfg.ContainerName
	if containerName == "" {
		containerName = generateName(cfg.NamePrefix)
	}
	if !dockerNameRegex.MatchString(containerName) {
		return fmt.Errorf("invalid container name: %q", containerName)
	}
	if nameExists(ctx, containerName) {
		return fmt.Errorf("container name already exists: %q", containerName)
	}

	// Validate host paths before docker run
	for i, p := range cfg.CopyPaths {
		if !filepath.IsAbs(p) {
			abs, _ := filepath.Abs(p)
			cfg.CopyPaths[i] = abs
		}
		resolved, err := filepath.EvalSymlinks(cfg.CopyPaths[i])
		if err == nil {
			cfg.CopyPaths[i] = resolved
		}
		if err := mustExist(cfg.CopyPaths[i]); err != nil {
			return fmt.Errorf("copy path invalid: %w", err)
		}
	}
	for i, m := range cfg.Mounts {
		host := m
		container := ""
		if strings.Contains(m, ":") {
			parts := strings.SplitN(m, ":", 2)
			host = parts[0]
			container = parts[1]
			if container != "" && !strings.HasPrefix(container, "/") {
				return fmt.Errorf("mount container path must be absolute: %q", container)
			}
		}
		if !filepath.IsAbs(host) {
			abs, _ := filepath.Abs(host)
			host = abs
		}
		if resolved, err := filepath.EvalSymlinks(host); err == nil {
			host = resolved
		}
		if err := mustExist(host); err != nil {
			return fmt.Errorf("mount host path invalid: %w", err)
		}
		// write back normalized mount string
		if container != "" {
			cfg.Mounts[i] = host + ":" + container
		} else {
			cfg.Mounts[i] = host
		}
	}

	// Ensure /tmp writable when read-only and no explicit writable paths
	if cfg.ReadOnlyRootFS && len(cfg.WritablePaths) == 0 {
		cfg.WritablePaths = []string{"/tmp"}
	}

	img := cfg.Image
	if img == "" {
		img = defaultImage
	}

	args := dockerargs.BuildRunArgs(cfg)
	args = append(args, "--detach")
	args = append(args, "--name", containerName)
	// keep container alive for exec/attach
	args = append(args, "--entrypoint", "tail")
	args = append(args, img, "-f", "/dev/null")

	if cfg.Verbose {
		fmt.Fprintln(os.Stderr, "+ docker", strings.Join(args, " "))
	}
	var runErr error
	if cfg.Verbose {
		runErr = runCmdInherit(ctx, "docker", args...)
	} else {
		runErr = runCmdQuiet(ctx, "docker", args...)
	}
	if runErr != nil {
		return fmt.Errorf("docker run failed: %w", runErr)
	}

	// Ensure cleanup on signals
	stopCh := make(chan os.Signal, 2)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stopCh)
	go func() {
		<-stopCh
		_ = gracefulStop(ctx, containerName, cfg)
	}()

	// Create writable paths inside container as root
	for _, w := range cfg.WritablePaths {
		_ = runCmdInherit(ctx, "docker", "exec", "-u", "0", containerName, "mkdir", "-p", w)
	}

	// Copy files (-c) into container user's home directory
	if len(cfg.CopyPaths) > 0 {
		homeDir, err := getContainerHome(ctx, containerName)
		if err != nil || homeDir == "" {
			// fallback sensible default for brew image
			homeDir = "/home/linuxbrew"
		}
		for _, host := range cfg.CopyPaths {
			if cfg.Verbose {
				fmt.Fprintln(os.Stderr, "+ docker cp", host, containerName+":"+homeDir)
			}
			if err := runCmdInherit(ctx, "docker", "cp", host, containerName+":"+homeDir); err != nil {
				return fmt.Errorf("docker cp %s failed: %w", host, err)
			}
			// make copied entry read-only (best-effort)
			base := filepath.Base(host)
			destPath := homeDir + "/" + base
			if cfg.Verbose {
				fmt.Fprintln(os.Stderr, "+ docker exec -u 0", containerName, "chmod -R a-w", destPath)
			}
			_ = runCmdInherit(ctx, "docker", "exec", "-u", "0", containerName, "sh", "-c", "chmod -R a-w "+shellQuote(destPath))
		}
	}

	// Install packages via brew if requested and network allowed
	if len(cfg.Packages) > 0 && !cfg.DisableNetwork {
		// Determine which packages are already installed and skip them
		filter := []string{}
		for _, p := range cfg.Packages {
			if !brewInstalled(ctx, containerName, p) {
				filter = append(filter, p)
			}
		}
		if len(filter) > 0 {
			installCmd := append([]string{"exec", containerName, "env",
				"HOMEBREW_NO_AUTO_UPDATE=1",
				"HOMEBREW_NO_INSTALL_CLEANUP=1",
				"HOMEBREW_NO_ENV_HINTS=1",
				"HOMEBREW_NO_ANALYTICS=1",
				"brew", "install", "-q"}, filter...)
			if cfg.Verbose {
				fmt.Fprintln(os.Stderr, "+ docker", strings.Join(installCmd, " "))
			}
			if err := runCmdInherit(ctx, "docker", installCmd...); err != nil {
				fmt.Fprintln(os.Stderr, "warning: brew install failed:", err)
			}
		}
	} else if len(cfg.Packages) > 0 && cfg.DisableNetwork {
		fmt.Fprintln(os.Stderr, "warning: --no-net set; skipping package installs")
	}

	// Ensure fish is available when preferred and network is allowed
	if (cfg.Shell == "" || cfg.Shell == "fish") && !cfg.DisableNetwork {
		if !hasShell(ctx, containerName, "fish") {
			cmd := []string{"exec", containerName, "env",
				"HOMEBREW_NO_AUTO_UPDATE=1",
				"HOMEBREW_NO_INSTALL_CLEANUP=1",
				"HOMEBREW_NO_ENV_HINTS=1",
				"HOMEBREW_NO_ANALYTICS=1",
				"brew", "install", "-q", "fish"}
			if cfg.Verbose {
				fmt.Fprintln(os.Stderr, "+ docker", strings.Join(cmd, " "))
			}
			_ = runCmdInherit(ctx, "docker", cmd...)
		}
	}

	// Open interactive shell based on preference (no screen clear)
	if err := execInteractiveShell(ctx, containerName, cfg.Shell, cfg.Verbose); err != nil {
		// even if shell fails, attempt cleanup below
		fmt.Fprintln(os.Stderr, "warning: interactive shell failed:", err)
	}

	// graceful stop and purge as configured
	if cfg.Verbose {
		if cfg.GracefulStop {
			fmt.Fprintln(os.Stderr, "+ docker stop --timeout 10", containerName)
		} else {
			fmt.Fprintln(os.Stderr, "+ docker kill", containerName)
		}
	}
	if err := gracefulStop(ctx, containerName, cfg); err != nil {
		return err
	}
	return nil
}

func mustExist(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func getContainerHome(ctx context.Context, container string) (string, error) {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "exec", container, "sh", "-lc", "printf %s \"$HOME\"")
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", err
	}
	home := strings.TrimSpace(out.String())
	if home == "" || !strings.HasPrefix(home, "/") {
		return "", fmt.Errorf("invalid HOME: %q", home)
	}
	return home, nil
}

func execInteractiveShell(ctx context.Context, container string, preferred string, verbose bool) error {
	// Build candidate names (not absolute paths) for probing
	names := []string{}
	if preferred != "" {
		names = append(names, preferred)
	}
	names = append(names, "fish", "bash", "zsh", "sh")
	var chosen string
	for _, name := range names {
		if hasShell(ctx, container, name) {
			chosen = name
			break
		}
	}
	if chosen == "" {
		return fmt.Errorf("no usable shell found")
	}
	args := []string{"exec", "-it", container, chosen}
	if verbose {
		fmt.Fprintln(os.Stderr, "+ docker", strings.Join(args, " "))
	}
	return runCmdAttach(ctx, "docker", args...)
}

func hasShell(ctx context.Context, container string, name string) bool {
	// Use sh -lc to run 'command -v <name>' quietly
	cmd := exec.CommandContext(ctx, "docker", "exec", container, "sh", "-lc", "command -v "+shellQuote(name)+" >/dev/null 2>&1")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func gracefulStop(ctx context.Context, container string, cfg config.RuntimeConfig) error {
	if cfg.GracefulStop {
		_ = runCmdInherit(ctx, "docker", "stop", "--timeout", "10", container)
	} else {
		_ = runCmdInherit(ctx, "docker", "kill", container)
	}
	if cfg.PurgeOnExit {
		_ = runCmdInherit(ctx, "docker", "rm", container)
	}
	return nil
}

func runCmdInherit(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runCmdQuiet(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runCmdAttach(ctx context.Context, name string, args ...string) error {
	// Same as inherit; separated for clarity
	return runCmdInherit(ctx, name, args...)
}

func generateName(prefix string) string {
	if prefix == "" {
		prefix = "lazylab"
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)
	return fmt.Sprintf("%s-%s-%05d", prefix, time.Now().Format("20060102150405"), rnd)
}

func shellQuote(s string) string {
	if strings.ContainsAny(s, " \"'$") {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "'\\''"))
	}
	return s
}

func nameExists(ctx context.Context, name string) bool {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	_ = cmd.Run()
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

func brewInstalled(ctx context.Context, container string, pkg string) bool {
	cmd := exec.CommandContext(ctx, "docker", "exec", container, "sh", "-lc", "brew list --formula --versions "+shellQuote(pkg)+" >/dev/null 2>&1")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
