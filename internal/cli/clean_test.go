package cli

import (
	"testing"

	"github.com/SomniSom/docker-ops/internal/config"
)

func TestImageRepository(t *testing.T) {
	if got := imageRepository("downloadbot:deploy"); got != "downloadbot" {
		t.Fatalf("got %q", got)
	}
	if got := imageRepository("ghcr.io/org/app:1.0.0"); got != "ghcr.io/org/app" {
		t.Fatalf("got %q", got)
	}
}

func TestProjectImageRefs(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		DeployImage:        "downloadbot:deploy",
		ComposeProjectName: "downloadbot",
		DeployMode:         "artifacts",
	}
	refs, err := projectImageRefs(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0] != "downloadbot:deploy" {
		t.Fatalf("refs=%v", refs)
	}
}
