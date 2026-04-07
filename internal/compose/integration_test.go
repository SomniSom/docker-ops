//go:build integration

package compose

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegration_DockerComposeVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("integration")
	}
	if err := LookPathDocker(); err != nil {
		t.Skip("docker not in PATH:", err)
	}
	dir := t.TempDir()
	compose := `services:
  sleep:
    image: busybox:latest
    command: ["sleep", "3600"]
`
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0o600); err != nil {
		t.Fatal(err)
	}
	r := &Runner{
		ProjectRoot:        dir,
		ComposeFile:        "docker-compose.yml",
		ComposeProjectName: "dqtest-" + filepath.Base(dir),
	}
	if err := r.Run("version"); err != nil {
		t.Fatal(err)
	}
}
