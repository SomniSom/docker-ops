package deploy

import (
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
)

// ModeIsArtifacts reports deploy_mode == artifacts (case-insensitive).
func ModeIsArtifacts(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.DeployMode), "artifacts")
}

func boolVal(p *bool) bool {
	return p != nil && *p
}

// ArtifactsUseSaveLoad matches the former bash deploy script’s registry vs save/load rules.
func ArtifactsUseSaveLoad(cfg *config.Config) bool {
	if cfg == nil {
		return true
	}
	if boolVal(cfg.DeploySaveLoad) {
		return true
	}
	if cfg.DeployUseRegistry != nil && !*cfg.DeployUseRegistry {
		return true
	}
	if cfg.DeployUseRegistry != nil && *cfg.DeployUseRegistry {
		return false
	}
	img := strings.TrimSpace(cfg.DeployImage)
	return !strings.Contains(img, "/")
}

// EffectiveSaveCompress defaults to true when unset (bash DEPLOY_SAVE_COMPRESS default 1).
func EffectiveSaveCompress(cfg *config.Config) bool {
	if cfg == nil || cfg.DeploySaveCompress == nil {
		return true
	}
	return *cfg.DeploySaveCompress
}

// EffectiveDeployPush is true when deploy_push is explicitly true.
func EffectiveDeployPush(cfg *config.Config) bool {
	return boolVal(cfg.DeployPush)
}
