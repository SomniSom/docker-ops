package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra/doc"

	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

func TestGenManRoot(t *testing.T) {
	locale.Set("en")
	root := NewRoot()
	var buf bytes.Buffer
	header := &doc.GenManHeader{Section: "1", Source: version.Version, Manual: version.Name}
	if err := doc.GenMan(root, header, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, ".TH") {
		t.Fatalf("expected groff .TH in output, got prefix %q", truncate(out, 200))
	}
}

func TestManFindDeploy(t *testing.T) {
	locale.Set("en")
	root := NewRoot()
	c, _, err := root.Find([]string{"deploy"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Name() != "deploy" {
		t.Fatalf("got %q", c.Name())
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
