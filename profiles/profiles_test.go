package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/isaaclins/LazyLab/config"
)

func TestSaveLoadList(t *testing.T) {
	// isolate profiles dir under temp HOME
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	if _, err := EnsureDir(); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	p := Profile{Name: "test", CPUs: "2", Memory: "1g"}
	if err := Save("test", p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load("test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.CPUs != "2" || got.Memory != "1g" {
		t.Fatalf("unexpected profile: %+v", got)
	}
	names, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 1 || names[0] != "test" {
		t.Fatalf("unexpected names: %v", names)
	}
	path := filepath.Join(dir, ".lazylab/profiles", "test.yaml")
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// perms 0600 expected; on some OS the umask may differ; accept owner-only writable
	if st.Mode().Perm()&0o077 != 0 {
		t.Fatalf("permissions too open: %v", st.Mode())
	}
}

func TestMergePrecedence(t *testing.T) {
	base := config.RuntimeConfig{CPULimit: "1", MemoryLimit: "512m", ReadOnlyRootFS: false}
	over := config.RuntimeConfig{CPULimit: "2", ReadOnlyRootFS: true}
	merged := config.MergeProfile(base, over)
	if merged.CPULimit != "2" {
		t.Fatalf("cpus not overridden: %+v", merged)
	}
	if merged.MemoryLimit != "512m" {
		t.Fatalf("memory changed unexpectedly: %+v", merged)
	}
	if !merged.ReadOnlyRootFS {
		t.Fatalf("readonly not set: %+v", merged)
	}
}
