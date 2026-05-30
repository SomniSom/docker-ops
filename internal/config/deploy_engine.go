package config

import "strings"

const (
	DeployEngineCompose = "compose"
	DeployEngineAuto    = "auto"
	DeployEngineAPI     = "api"
)

// EffectiveDeployEngine returns compose when unset (backward compatible).
func EffectiveDeployEngine(c *Config) string {
	if c == nil {
		return DeployEngineCompose
	}
	e := strings.TrimSpace(strings.ToLower(c.DeployEngine))
	switch e {
	case "", DeployEngineCompose:
		return DeployEngineCompose
	case DeployEngineAuto, DeployEngineAPI:
		return e
	default:
		return DeployEngineCompose
	}
}

// UsesDeployAPI reports whether artifacts deploy may use Docker Engine API paths.
func UsesDeployAPI(c *Config) bool {
	switch EffectiveDeployEngine(c) {
	case DeployEngineAuto, DeployEngineAPI:
		return true
	default:
		return false
	}
}

// EffectiveDeploySkipUnchanged defaults to true when unset.
func EffectiveDeploySkipUnchanged(c *Config) bool {
	if c == nil || c.DeploySkipUnchanged == nil {
		return true
	}
	return *c.DeploySkipUnchanged
}

// EffectiveDeployLayerSync defaults to true when unset.
func EffectiveDeployLayerSync(c *Config) bool {
	if c == nil || c.DeployLayerSync == nil {
		return true
	}
	return *c.DeployLayerSync
}

// RemoteDockerSocketExplicit reports a non-auto explicit unix socket path from config.
func RemoteDockerSocketExplicit(c *Config) (path string, ok bool) {
	if c == nil {
		return "", false
	}
	s := strings.TrimSpace(c.RemoteDockerSocket)
	if s == "" || strings.EqualFold(s, "auto") {
		return "", false
	}
	return s, true
}
