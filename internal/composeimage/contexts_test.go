package composeimage

import (
	"path/filepath"
	"testing"
)

func TestServiceBuildDirs(t *testing.T) {
	root := t.TempDir()
	yml := `services:
  app:
    build: .
  worker:
    build:
      context: ./worker
`
	m, err := ServiceBuildDirs([]byte(yml), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 2 {
		t.Fatalf("got %d services", len(m))
	}
	if a := m["app"]; filepath.Clean(a) != filepath.Clean(root) {
		t.Fatalf("app context %q want %q", a, root)
	}
	if w := m["worker"]; filepath.Base(w) != "worker" {
		t.Fatalf("worker context %q", w)
	}
}
