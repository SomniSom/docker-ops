package imagecmp

import (
	"context"
	"testing"

	"github.com/SomniSom/docker-ops/internal/dockerapi"
)

func TestFilterUnchangedSkipDisabled(t *testing.T) {
	refs := []string{"a:1", "b:2"}
	need, skipped, err := FilterUnchanged(context.Background(), nil, nil, refs, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 0 || len(need) != 2 {
		t.Fatalf("got need=%v skipped=%v", need, skipped)
	}
}

func TestFilterUnchangedFailOpenOnInspectError(t *testing.T) {
	need, skipped, err := FilterUnchanged(context.Background(), &dockerapi.Client{}, &dockerapi.Client{}, []string{"x"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(need) != 1 || len(skipped) != 0 {
		t.Fatalf("fail-open: need=%v skipped=%v", need, skipped)
	}
}
