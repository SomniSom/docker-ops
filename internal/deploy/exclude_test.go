package deploy

import (
	"testing"

	"github.com/SomniSom/docker-ops/internal/config"
)

func TestPathExcluded(t *testing.T) {
	patterns := []string{".git/", "data/", "*.db", ".env", "dq.env"}
	tests := []struct {
		rel  string
		want bool
	}{
		{".git/config", true},
		{"pkg/.git", true},
		{"src/main.go", false},
		{"data/x", true},
		{"foo/data/bar", true},
		{"x.db", true},
		{"dir/x.db", true},
		{".env", true},
		{"dq.env", true},
		{"sub/.env", true},
	}
	for _, tt := range tests {
		if got := PathExcluded(tt.rel, patterns); got != tt.want {
			t.Errorf("PathExcluded(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}

func TestArtifactsUseSaveLoad(t *testing.T) {
	regFalse := false
	t.Run("registry off uses save load", func(t *testing.T) {
		cfg := &config.Config{DeployImage: "localtag", DeployUseRegistry: &regFalse}
		if !ArtifactsUseSaveLoad(cfg) {
			t.Fatal("expected save/load when deploy_use_registry is false")
		}
	})
	t.Run("auto local tag", func(t *testing.T) {
		cfg := &config.Config{DeployImage: "myapp:1"}
		if !ArtifactsUseSaveLoad(cfg) {
			t.Fatal("expected save/load for tag without slash")
		}
	})
	t.Run("auto registry path", func(t *testing.T) {
		cfg := &config.Config{DeployImage: "registry.io/ns/app:1"}
		if ArtifactsUseSaveLoad(cfg) {
			t.Fatal("expected registry for namespaced image")
		}
	})
}
