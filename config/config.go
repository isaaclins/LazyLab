package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RuntimeConfig represents the full set of options to run a container
// Naming uses verbose, explicit fields for clarity

type RuntimeConfig struct {
	Packages        []string
	CopyPaths       []string
	Mounts          []string
	ContainerName   string
	NamePrefix      string
	PurgeOnExit     bool
	DisableNetwork  bool
	MemoryLimit     string
	CPULimit        string
	PidsLimit       int
	ReadOnlyRootFS  bool
	WritablePaths   []string
	ForceAMD64      bool
	GracefulStop    bool
	ProfileName     string
	Image           string
	Verbose         bool
	CapDropAll      bool
	NoNewPrivileges bool
	Shell           string
	CachePackages   bool
	User            string
	PurgeCache      bool
}

// FromCommand builds the config from cobra flags (profile merge not yet applied)
func FromCommand(cmd *cobra.Command) (RuntimeConfig, error) {
	var rc RuntimeConfig
	var err error

	rc.Packages, _ = cmd.Flags().GetStringSlice("packages")
	rc.CopyPaths, _ = cmd.Flags().GetStringSlice("copy")
	rc.Mounts, _ = cmd.Flags().GetStringSlice("mount")
	rc.ContainerName, _ = cmd.Flags().GetString("name")
	rc.NamePrefix, _ = cmd.Flags().GetString("prefix")
	rc.PurgeOnExit, _ = cmd.Flags().GetBool("purge")
	rc.DisableNetwork, _ = cmd.Flags().GetBool("no-net")
	rc.MemoryLimit, _ = cmd.Flags().GetString("memory")
	rc.CPULimit, _ = cmd.Flags().GetString("cpus")
	rc.PidsLimit, _ = cmd.Flags().GetInt("pids-limit")
	rc.ReadOnlyRootFS, _ = cmd.Flags().GetBool("read-only")
	rc.WritablePaths, _ = cmd.Flags().GetStringSlice("writable")
	rc.ForceAMD64, _ = cmd.Flags().GetBool("amd64")
	rc.GracefulStop, _ = cmd.Flags().GetBool("graceful")
	rc.ProfileName, _ = cmd.Flags().GetString("profile")
	rc.Image, _ = cmd.Flags().GetString("image")
	rc.Verbose, _ = cmd.Flags().GetBool("verbose")
	rc.CapDropAll, _ = cmd.Flags().GetBool("cap-drop-all")
	rc.NoNewPrivileges, _ = cmd.Flags().GetBool("no-new-privs")
	rc.Shell, _ = cmd.Flags().GetString("shell")
	rc.CachePackages, _ = cmd.Flags().GetBool("cache-packages")
	rc.User, _ = cmd.Flags().GetString("user")
	rc.PurgeCache, _ = cmd.Flags().GetBool("purge-cache")

	if err != nil {
		return rc, err
	}

	// Basic validation placeholders; detailed validation to be added
	if rc.ReadOnlyRootFS && len(rc.WritablePaths) == 0 {
		// Not an error, but recommend defaults later
	}
	if rc.DisableNetwork && len(rc.Packages) > 0 {
		// Will warn/skip installs at runtime
	}

	return rc, nil
}

// MergeProfile merges profile-derived config (base) with CLI-derived config (overrides)
func MergeProfile(base RuntimeConfig, override RuntimeConfig) RuntimeConfig {
	merged := base
	// Simple precedence: if override has non-zero value, apply
	if len(override.Packages) > 0 {
		merged.Packages = override.Packages
	}
	if len(override.CopyPaths) > 0 {
		merged.CopyPaths = override.CopyPaths
	}
	if len(override.Mounts) > 0 {
		merged.Mounts = override.Mounts
	}
	if override.ContainerName != "" {
		merged.ContainerName = override.ContainerName
	}
	if override.NamePrefix != "" {
		merged.NamePrefix = override.NamePrefix
	}
	if override.PurgeOnExit {
		merged.PurgeOnExit = true
	}
	if override.DisableNetwork {
		merged.DisableNetwork = true
	}
	if override.MemoryLimit != "" {
		merged.MemoryLimit = override.MemoryLimit
	}
	if override.CPULimit != "" {
		merged.CPULimit = override.CPULimit
	}
	if override.PidsLimit != 0 {
		merged.PidsLimit = override.PidsLimit
	}
	if override.ReadOnlyRootFS {
		merged.ReadOnlyRootFS = true
	}
	if len(override.WritablePaths) > 0 {
		merged.WritablePaths = override.WritablePaths
	}
	if override.ForceAMD64 {
		merged.ForceAMD64 = true
	}
	if override.Image != "" {
		merged.Image = override.Image
	}
	if override.Verbose {
		merged.Verbose = true
	}
	if override.CapDropAll {
		merged.CapDropAll = true
	}
	if override.NoNewPrivileges {
		merged.NoNewPrivileges = true
	}
	if override.Shell != "" {
		merged.Shell = override.Shell
	}
	merged.CachePackages = override.CachePackages || base.CachePackages
	if override.User != "" {
		merged.User = override.User
	}
	merged.PurgeCache = override.PurgeCache || base.PurgeCache
	// GracefulStop defaults to false now; only override when explicitly true via flag/profile
	return merged
}

// String returns a brief summary for display
func (r RuntimeConfig) String() string {
	return fmt.Sprintf("name=%s prefix=%s net=%v ro=%v cpus=%s mem=%s image=%s",
		r.ContainerName, r.NamePrefix, !r.DisableNetwork, r.ReadOnlyRootFS, r.CPULimit, r.MemoryLimit, r.Image,
	)
}
