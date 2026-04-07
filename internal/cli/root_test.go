package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SomniSom/docker-ops/internal/version"
)

func TestNewRoot_Version(t *testing.T) {
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte(version.Name)) {
		t.Fatalf("output missing product name: %q", buf.String())
	}
}

func TestNewRoot_EnvAnonymize(t *testing.T) {
	root := NewRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"env", "--anonymize", "--project-dir", t.TempDir()})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "user@host") {
		t.Fatalf("expected placeholder: %s", out)
	}
}
