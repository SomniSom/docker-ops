package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckAppConfigExists(t *testing.T) {
	dir := t.TempDir()
	if err := checkAppConfigExists(dir, ""); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(f, []byte("k: v\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkAppConfigExists(dir, "app.yaml"); err != nil {
		t.Fatal(err)
	}
	if err := checkAppConfigExists(dir, "missing.yaml"); err == nil {
		t.Fatal("expected error")
	}
}
