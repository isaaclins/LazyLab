package profiles

import (
	"os"
	"testing"
)

func TestJSONProfileSaveLoadAndList(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	if _, err := EnsureDir(); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	p := Profile{Name: "demo", Image: "ubuntu:22.04", Purge: true}
	if err := Save("demo.json", p); err != nil {
		t.Fatalf("Save json: %v", err)
	}
	got, err := Load("demo")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Image != "ubuntu:22.04" || !got.Purge {
		t.Fatalf("unexpected: %+v", got)
	}
	names, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 1 || names[0] != "demo" {
		t.Fatalf("names: %v", names)
	}
}
