package deploy

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/opencontainers/go-digest"
)

// serviceConfigHash matches docker compose ServiceHash (config --hash) so API-created
// containers are visible to docker compose ps/logs/exec.
func serviceConfigHash(svc types.ServiceConfig) (string, error) {
	svc.Build = nil
	svc.PullPolicy = ""
	svc.Scale = nil
	if svc.Deploy != nil {
		svc.Deploy.Replicas = nil
	}
	svc.DependsOn = nil
	svc.Profiles = nil

	b, err := json.Marshal(svc)
	if err != nil {
		return "", err
	}
	return digest.SHA256.FromBytes(b).Encoded(), nil
}

func composeDependsOnLabel(svc types.ServiceConfig) string {
	var deps []string
	for s, d := range svc.DependsOn {
		deps = append(deps, fmt.Sprintf("%s:%s:%t", s, d.Condition, d.Restart))
	}
	return strings.Join(deps, ",")
}

func composeConfigFilesLabel(cfg *config.Config, projectRoot, composeFile string) string {
	base := filepath.Base(composeFile)
	if base == "" {
		base = config.ArtifactsComposeFileName
	}
	if rp := strings.TrimSpace(cfg.RemotePath); cfg != nil && cfg.RemoteConfigured() && rp != "" {
		return remoteJoin(rp, base)
	}
	if filepath.IsAbs(composeFile) {
		return composeFile
	}
	if projectRoot != "" {
		return filepath.Join(projectRoot, base)
	}
	return base
}
