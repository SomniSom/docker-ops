package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEffectiveComposeFileSourceMode(t *testing.T) {
	c := &Config{DeployMode: "source", ComposeFile: "docker-compose.yml"}
	if got := EffectiveComposeFile(c, t.TempDir()); got != "docker-compose.yml" {
		t.Fatalf("got %q", got)
	}
}

func TestEffectiveComposeFileArtifactsWithImageCompose(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ArtifactsComposeFileName), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &Config{DeployMode: "artifacts", ComposeFile: "docker-compose.yml"}
	if got := EffectiveComposeFile(c, dir); got != ArtifactsComposeFileName {
		t.Fatalf("got %q want %q", got, ArtifactsComposeFileName)
	}
}

func TestEffectiveComposeFileArtifactsExplicitImage(t *testing.T) {
	c := &Config{DeployMode: "artifacts", ComposeFile: ArtifactsComposeFileName}
	if got := EffectiveComposeFile(c, t.TempDir()); got != ArtifactsComposeFileName {
		t.Fatalf("got %q", got)
	}
}
