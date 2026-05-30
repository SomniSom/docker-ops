package deploy

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestRestartPolicyFromString(t *testing.T) {
	tests := []struct {
		in   string
		want container.RestartPolicyMode
	}{
		{"always", container.RestartPolicyAlways},
		{"unless-stopped", container.RestartPolicyUnlessStopped},
		{"", container.RestartPolicyDisabled},
	}
	for _, tc := range tests {
		got := restartPolicyFromString(tc.in)
		if got.Name != tc.want {
			t.Errorf("%q => %q, want %q", tc.in, got.Name, tc.want)
		}
	}
}

func TestMappingToEnv(t *testing.T) {
	// smoke: empty mapping
	if len(mappingToEnv(nil)) != 0 {
		t.Fatal("expected empty env")
	}
}
