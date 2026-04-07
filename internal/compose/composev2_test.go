package compose

import (
	"errors"
	"strings"
	"testing"

	"github.com/SomniSom/docker-ops/internal/locale"
)

func TestLooksLikeMissingComposePlugin(t *testing.T) {
	cases := []struct {
		stderr string
		want   bool
	}{
		{"docker: 'compose' is not a docker command.", true},
		{"docker: compose is not a docker command.", true},
		{"unknown command: compose", true},
		{"Cannot connect to the Docker daemon", false},
		{"no configuration file provided", false},
	}
	for _, tc := range cases {
		if got := looksLikeMissingComposePlugin(tc.stderr); got != tc.want {
			t.Errorf("%q: got %v want %v", tc.stderr, got, tc.want)
		}
	}
}

func TestHintIfComposePluginError(t *testing.T) {
	locale.Set("en")
	base := errors.New("exit status 1")
	err := HintIfComposePluginError("docker: 'compose' is not a docker command.", base)
	if err == base {
		t.Fatal("expected wrapped error with hint")
	}
	s := err.Error()
	if !strings.Contains(s, "docker-compose-plugin") && !strings.Contains(s, "Compose V2") {
		t.Fatalf("expected install hint: %s", s)
	}
}
