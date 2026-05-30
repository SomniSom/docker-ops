package config

import "testing"

func TestEffectiveDeployEngineDefaultCompose(t *testing.T) {
	if got := EffectiveDeployEngine(nil); got != DeployEngineCompose {
		t.Fatalf("nil cfg: got %q", got)
	}
	if got := EffectiveDeployEngine(&Config{}); got != DeployEngineCompose {
		t.Fatalf("empty cfg: got %q", got)
	}
}

func TestEffectiveDeployEngineValues(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"auto", DeployEngineAuto},
		{"api", DeployEngineAPI},
		{"compose", DeployEngineCompose},
		{"invalid", DeployEngineCompose},
	}
	for _, tc := range tests {
		got := EffectiveDeployEngine(&Config{DeployEngine: tc.in})
		if got != tc.want {
			t.Errorf("DeployEngine %q => %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEffectiveDeploySkipUnchangedDefault(t *testing.T) {
	if !EffectiveDeploySkipUnchanged(nil) {
		t.Fatal("expected default true")
	}
}

func TestUsesDeployAPI(t *testing.T) {
	if UsesDeployAPI(&Config{}) {
		t.Fatal("compose default should not use API path for finish")
	}
	if !UsesDeployAPI(&Config{DeployEngine: "auto"}) {
		t.Fatal("auto should use API")
	}
}
