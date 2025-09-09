package dockerargs

import (
	"strings"
	"testing"

	"github.com/isaaclins/LazyLab/config"
)

func TestBuildRunArgs_UserAndMountDefault(t *testing.T) {
	cfg := config.RuntimeConfig{
		User:        "1000:1000",
		Mounts:      []string{"/host/path"},
		Image:       "homebrew/brew:latest",
		PurgeOnExit: true,
	}
	args := BuildRunArgs(cfg)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--user 1000:1000") {
		t.Fatalf("expected --user in args: %s", joined)
	}
	if !strings.Contains(joined, "type=bind,src=/host/path,dst=/home/linuxbrew/path") {
		t.Fatalf("expected default home mount: %s", joined)
	}
}
