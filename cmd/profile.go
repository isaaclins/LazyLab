package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/isaaclins/LazyLab/config"
	"github.com/isaaclins/LazyLab/profiles"
	"github.com/isaaclins/LazyLab/runner"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage lazylab profiles",
}

var profileSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save current flags as a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		rc, err := config.FromCommand(cmd.Root())
		if err != nil {
			return err
		}
		p := profiles.Profile{
			Name:          name,
			Packages:      rc.Packages,
			Copy:          rc.CopyPaths,
			Mounts:        rc.Mounts,
			ContainerName: rc.ContainerName,
			Prefix:        rc.NamePrefix,
			Purge:         rc.PurgeOnExit,
			NoNet:         rc.DisableNetwork,
			Memory:        rc.MemoryLimit,
			CPUs:          rc.CPULimit,
			PidsLimit:     rc.PidsLimit,
			ReadOnly:      rc.ReadOnlyRootFS,
			Writable:      rc.WritablePaths,
			AMD64:         rc.ForceAMD64,
			Graceful:      rc.GracefulStop,
		}
		if err := profiles.Save(name, p); err != nil {
			return err
		}
		fmt.Printf("Saved profile %s\n", name)
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := profiles.List()
		if err != nil {
			return err
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	},
}

var profileRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run container using a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		p, err := profiles.Load(name)
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
		}
		override, err := config.FromCommand(cmd.Root())
		if err != nil {
			return err
		}
		merged := config.MergeProfile(base, override)
		ctx := context.Background()
		return runner.Run(ctx, merged)
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := profiles.Delete(name); err != nil {
			return err
		}
		fmt.Printf("Deleted profile %s\n", name)
		return nil
	},
}

var profileEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a profile in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		editor, _ := cmd.Flags().GetString("editor")
		path, err := profiles.OpenInEditor(name, editor)
		if err != nil {
			return err
		}
		c := exec.Command(editor, path)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileSaveCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileRunCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileEditCmd)

	// Allow passing flags to profile run to override
	profileRunCmd.Flags().AddFlagSet(rootCmd.PersistentFlags())

	// Helpful env for editor
	profileEditCmd.Flags().String("editor", os.Getenv("EDITOR"), "Editor to use for editing the profile")
}
