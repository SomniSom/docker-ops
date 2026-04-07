package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunner_Command_Args(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services: {}\n"), 0o600)

	r := &Runner{
		ProjectRoot:        dir,
		ComposeFile:        "docker-compose.yml",
		ComposeProjectName: "p1",
	}
	cmd := r.Command("ps")
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "compose") || !strings.Contains(args, "p1") || !strings.Contains(args, "docker-compose.yml") {
		t.Fatalf("unexpected args: %v", cmd.Args)
	}
	if cmd.Dir != dir {
		t.Fatalf("Dir: %s", cmd.Dir)
	}
}

func TestRunner_Run_MissingComposeFile(t *testing.T) {
	dir := t.TempDir()
	r := &Runner{ProjectRoot: dir, ComposeFile: "missing.yml", ComposeProjectName: "x"}
	err := r.Run("ps")
	if err == nil {
		t.Fatal("expected error")
	}
}
