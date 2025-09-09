package dockerargs

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/isaaclins/LazyLab/config"
)

// BuildRunArgs converts RuntimeConfig into docker run arguments (no exec)
func BuildRunArgs(cfg config.RuntimeConfig) []string {
	args := []string{"run"}
	if cfg.PurgeOnExit {
		args = append(args, "--rm")
	}
	if cfg.DisableNetwork {
		args = append(args, "--network", "none")
	}
	if cfg.ReadOnlyRootFS {
		args = append(args, "--read-only")
	}
	for _, w := range cfg.WritablePaths {
		args = append(args, "--mount", fmt.Sprintf("type=tmpfs,destination=%s", w))
	}
	if cfg.MemoryLimit != "" {
		args = append(args, "--memory", cfg.MemoryLimit)
	}
	if cfg.CPULimit != "" {
		args = append(args, "--cpus", cfg.CPULimit)
	}
	if cfg.PidsLimit != 0 {
		args = append(args, "--pids-limit", fmt.Sprintf("%d", cfg.PidsLimit))
	}
	if cfg.ForceAMD64 {
		args = append(args, "--platform", "linux/amd64")
	}
	if cfg.CapDropAll {
		args = append(args, "--cap-drop", "ALL")
	}
	if cfg.NoNewPrivileges {
		args = append(args, "--security-opt", "no-new-privileges:true")
	}
	if cfg.User != "" {
		args = append(args, "--user", cfg.User)
	}
	defaultHome := guessHomeForImage(cfg.Image)
	for _, m := range cfg.Mounts {
		// accept host[:container]
		host := m
		container := ""
		if strings.Contains(m, ":") {
			parts := strings.SplitN(m, ":", 2)
			host = parts[0]
			container = parts[1]
		} else {
			// default destination under container home
			container = filepath.ToSlash(filepath.Join(defaultHome, filepath.Base(host)))
		}
		args = append(args, "--mount", fmt.Sprintf("type=bind,src=%s,dst=%s", host, container))
	}
	if cfg.ContainerName != "" {
		args = append(args, "--name", cfg.ContainerName)
	}
	// image placeholder; to be provided by caller
	return args
}

func guessHomeForImage(image string) string {
	if image == "" || strings.Contains(image, "homebrew/brew") {
		return "/home/linuxbrew"
	}
	return "/root"
}
