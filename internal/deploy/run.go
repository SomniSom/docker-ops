package deploy

import (
	"fmt"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
)

// RunOpts carries CLI-only options for deploy (not read from docker-ops.yml).
type RunOpts struct {
	// Build runs docker build -t deploy_image before save/load or registry push.
	Build bool
}

// Run runs deploy in source or artifacts mode per cfg.DeployMode.
func Run(projectRoot string, cfg *config.Config) error {
	return RunWithOptions(projectRoot, cfg, RunOpts{})
}

// RunWithOptions is like Run but honors CLI flags in opts (e.g. Build).
func RunWithOptions(projectRoot string, cfg *config.Config, opts RunOpts) error {
	if cfg == nil {
		return fmt.Errorf("%s", locale.T("err.config_required"))
	}
	if ModeIsArtifacts(cfg) {
		return RunArtifacts(projectRoot, cfg, opts)
	}
	return RunSource(projectRoot, cfg)
}
