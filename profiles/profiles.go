package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	profilesDirName = ".lazylab/profiles"
)

// Profile mirrors the schema described in rules

type Profile struct {
	Name          string   `yaml:"name" json:"name"`
	Image         string   `yaml:"image" json:"image"`
	Packages      []string `yaml:"packages" json:"packages"`
	Copy          []string `yaml:"copy" json:"copy"`
	Mounts        []string `yaml:"mounts" json:"mounts"`
	ContainerName string   `yaml:"containerName" json:"containerName"`
	Prefix        string   `yaml:"prefix" json:"prefix"`
	Purge         bool     `yaml:"purge" json:"purge"`
	NoNet         bool     `yaml:"noNet" json:"noNet"`
	Memory        string   `yaml:"memory" json:"memory"`
	CPUs          string   `yaml:"cpus" json:"cpus"`
	PidsLimit     int      `yaml:"pidsLimit" json:"pidsLimit"`
	ReadOnly      bool     `yaml:"readOnly" json:"readOnly"`
	Writable      []string `yaml:"writable" json:"writable"`
	AMD64         bool     `yaml:"amd64" json:"amd64"`
	Graceful      bool     `yaml:"graceful" json:"graceful"`
}

func ProfilesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, profilesDirName), nil
}

func EnsureDir() (string, error) {
	dir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func sanitizeName(name string) error {
	if name == "" {
		return errors.New("empty profile name")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		return errors.New("invalid profile name")
	}
	return nil
}

func pathFor(name string) (string, error) {
	if err := sanitizeName(name); err != nil {
		return "", err
	}
	dir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".yaml"), nil
}

func Save(name string, p Profile) error {
	if err := sanitizeName(name); err != nil {
		return err
	}
	dir, err := EnsureDir()
	if err != nil {
		return err
	}
	var path string
	// preserve extension if exists
	if strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
		path = filepath.Join(dir, name)
	} else {
		path = filepath.Join(dir, name+".yaml")
	}
	var b []byte
	if strings.HasSuffix(path, ".json") {
		b, err = json.MarshalIndent(p, "", "  ")
	} else {
		b, err = yaml.Marshal(p)
	}
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return nil
}

func Load(name string) (Profile, error) {
	var p Profile
	path, err := pathFor(name)
	if err != nil {
		return p, err
	}
	// try json/yaml variants
	candidates := []string{path}
	if !strings.HasSuffix(path, ".json") {
		candidates = append([]string{strings.TrimSuffix(path, ".yaml") + ".json"}, candidates...)
	}
	var data []byte
	for _, cand := range candidates {
		if b, err := os.ReadFile(cand); err == nil {
			data = b
			path = cand
			break
		}
	}
	if len(data) == 0 {
		return p, fmt.Errorf("profile not found: %s", name)
	}
	if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(data, &p); err != nil {
			return p, err
		}
	} else {
		if err := yaml.Unmarshal(data, &p); err != nil {
			return p, err
		}
	}
	return p, nil
}

func Delete(name string) error {
	path, err := pathFor(name)
	if err != nil {
		return err
	}
	// remove both yaml/json if present
	_ = os.Remove(strings.TrimSuffix(path, ".yaml") + ".json")
	return os.Remove(path)
}

func List() ([]string, error) {
	dir, err := ProfilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".json"))
		}
	}
	return names, nil
}

func OpenInEditor(name string, editor string) (string, error) {
	path, err := pathFor(name)
	if err != nil {
		return "", err
	}
	if editor == "" {
		return path, fmt.Errorf("no editor specified")
	}
	return path, nil
}
