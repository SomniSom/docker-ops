package config

import (
	"os"
	"strconv"
	"strings"
)

// applyDQEnvMap merges dq.env into c (§14.1 step 2). Only non-empty values from the map.
func applyDQEnvMap(c *Config, m map[string]string) {
	if c == nil || len(m) == 0 {
		return
	}
	set := func(key string, dest *string) {
		if v, ok := m[key]; ok && strings.TrimSpace(v) != "" {
			*dest = strings.TrimSpace(v)
		}
	}
	set("COMPOSE_PROJECT_NAME", &c.ComposeProjectName)
	set("COMPOSE_FILE", &c.ComposeFile)
	set("COMPOSE_SERVICE", &c.ComposeService)
	set("REMOTE_SSH", &c.RemoteSSH)
	set("REMOTE_PATH", &c.RemotePath)
	set("SSH_IDENTITY", &c.SSHIdentity)
	set("RSYNC_EXTRA", &c.RsyncExtra)
	set("DEPLOY_MODE", &c.DeployMode)
	set("DEPLOY_IMAGE", &c.DeployImage)
	set("APP_CONFIG", &c.AppConfig)

	if v, ok := m["DEPLOY_PUSH"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.DeployPush = boolPtr(b)
		}
	}
	if v, ok := m["DEPLOY_USE_REGISTRY"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.DeployUseRegistry = boolPtr(b)
		}
	}
	if v, ok := m["DEPLOY_SAVE_LOAD"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.DeploySaveLoad = boolPtr(b)
		}
	}
	if v, ok := m["DEPLOY_USE_SAVE_LOAD"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.DeploySaveLoad = boolPtr(b)
		}
	}
	if v, ok := m["DEPLOY_SAVE_COMPRESS"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.DeploySaveCompress = boolPtr(b)
		}
	}
	if v, ok := m["DEPLOY_BUILD_REMOTE"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.DeployBuildRemote = boolPtr(b)
		}
	}
	if v, ok := m["DOCKER_OPS_USE_REMOTE"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.UseRemote = boolPtr(b)
		}
	}
	if v, ok := m["DOCKER_OPS_HELP_SHOW_EFFECTIVE"]; ok {
		if b, ok := parseBoolString(v); ok {
			c.HelpShowEffective = boolPtr(b)
		}
	}
}

// overlayProcessEnv applies process environment (§14.1 step 3 — highest priority).
func overlayProcessEnv(c *Config) {
	if c == nil {
		return
	}
	set := func(key string, dest *string) {
		if v := os.Getenv(key); v != "" {
			*dest = v
		}
	}
	set("COMPOSE_PROJECT_NAME", &c.ComposeProjectName)
	set("COMPOSE_FILE", &c.ComposeFile)
	set("COMPOSE_SERVICE", &c.ComposeService)
	set("REMOTE_SSH", &c.RemoteSSH)
	set("REMOTE_PATH", &c.RemotePath)
	set("SSH_IDENTITY", &c.SSHIdentity)
	set("RSYNC_EXTRA", &c.RsyncExtra)
	set("DEPLOY_MODE", &c.DeployMode)
	set("DEPLOY_IMAGE", &c.DeployImage)
	set("APP_CONFIG", &c.AppConfig)

	if v := os.Getenv("DEPLOY_PUSH"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.DeployPush = boolPtr(b)
		}
	}
	if v := os.Getenv("DEPLOY_USE_REGISTRY"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.DeployUseRegistry = boolPtr(b)
		}
	}
	if v := os.Getenv("DEPLOY_SAVE_LOAD"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.DeploySaveLoad = boolPtr(b)
		}
	}
	if v := os.Getenv("DEPLOY_USE_SAVE_LOAD"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.DeploySaveLoad = boolPtr(b)
		}
	}
	if v := os.Getenv("DEPLOY_SAVE_COMPRESS"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.DeploySaveCompress = boolPtr(b)
		}
	}
	if v := os.Getenv("DEPLOY_BUILD_REMOTE"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.DeployBuildRemote = boolPtr(b)
		}
	}
	if v := os.Getenv("DOCKER_OPS_USE_REMOTE"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.UseRemote = boolPtr(b)
		}
	}
	if v := os.Getenv("DOCKER_OPS_HELP_SHOW_EFFECTIVE"); v != "" {
		if b, ok := parseBoolString(v); ok {
			c.HelpShowEffective = boolPtr(b)
		}
	}
}

func parseBoolString(s string) (bool, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			if i == 0 {
				return false, true
			}
			if i == 1 {
				return true, true
			}
		}
	}
	return false, false
}

func boolPtr(b bool) *bool { return &b }
