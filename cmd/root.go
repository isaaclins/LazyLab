package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/isaaclins/LazyLab/config"
	"github.com/isaaclins/LazyLab/profiles"
	"github.com/isaaclins/LazyLab/runner"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "lazylab",
		Short:   "Run and manage malware analysis containers",
		Long:    "lazylab simplifies running Docker containers for malware analysis with resource limits, isolation, and profiles.",
		Version: fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date),
		RunE: func(cmd *cobra.Command, args []string) error {
			rc, err := config.FromCommand(cmd)
			if err != nil {
				return err
			}
			if rc.ProfileName != "" {
				p, err := profiles.Load(rc.ProfileName)
				if err != nil {
					return err
				}
				base := config.RuntimeConfig{
					Packages:       p.Packages,
					CopyPaths:      p.Copy,
					Mounts:         p.Mounts,
					ContainerName:  p.ContainerName,
					NamePrefix:     p.Prefix,
					PurgeOnExit:    p.Purge,
					DisableNetwork: p.NoNet,
					MemoryLimit:    p.Memory,
					CPULimit:       p.CPUs,
					PidsLimit:      p.PidsLimit,
					ReadOnlyRootFS: p.ReadOnly,
					WritablePaths:  p.Writable,
					ForceAMD64:     p.AMD64,
					GracefulStop:   p.Graceful,
					Image:          p.Image,
					Shell:          p.Name, // fallback if profile adds shell later
				}
				rc = config.MergeProfile(base, rc)
			}
			ctx := context.Background()
			return runner.Run(ctx, rc)
		},
	}
)

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags (subset; full parsing added with config package)
	rootCmd.PersistentFlags().StringSliceP("packages", "p", nil, "Packages to install via brew inside container")
	rootCmd.PersistentFlags().StringSliceP("copy", "c", nil, "Copy files/dirs into container (immutable)")
	rootCmd.PersistentFlags().StringSliceP("mount", "m", nil, "Bind mount files/dirs (host[:container])")
	rootCmd.PersistentFlags().StringP("name", "n", "", "Custom container name")
	rootCmd.PersistentFlags().String("prefix", "", "Prefix for generated container names")
	rootCmd.PersistentFlags().Bool("purge", false, "Stop and remove container after exit")
	rootCmd.PersistentFlags().Bool("no-net", false, "Disable network access inside container")
	rootCmd.PersistentFlags().String("memory", "", "Limit container memory (e.g., 1g)")
	rootCmd.PersistentFlags().String("cpus", "", "Limit number of CPUs")
	rootCmd.PersistentFlags().Int("pids-limit", 0, "Limit number of processes inside container")
	rootCmd.PersistentFlags().Bool("read-only", false, "Make container filesystem read-only")
	rootCmd.PersistentFlags().StringSlice("writable", nil, "Paths to remain writable in read-only mode")
	rootCmd.PersistentFlags().Bool("amd64", false, "Force container architecture to amd64")
	rootCmd.PersistentFlags().Bool("graceful", false, "Gracefully stop the container with cleanup")
	rootCmd.PersistentFlags().String("profile", "", "Load and merge the named profile")
	rootCmd.PersistentFlags().String("image", "", "Container image to use (default homebrew/brew:latest)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Print docker commands and debug info")
	rootCmd.PersistentFlags().Bool("cap-drop-all", false, "Drop all Linux capabilities in container")
	rootCmd.PersistentFlags().Bool("no-new-privs", false, "Set no-new-privileges security option")
	rootCmd.PersistentFlags().String("shell", "fish", "Shell to start inside container (fish, bash, zsh, sh)")
	rootCmd.PersistentFlags().Bool("cache-packages", true, "Cache package manager downloads and reuse across runs")
	rootCmd.PersistentFlags().String("user", "", "Run container processes as user (e.g., 1000:1000 or name)")
	rootCmd.PersistentFlags().Bool("purge-cache", false, "Remove lazylab cache volume on exit")
}
