package composeimage

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ServiceBuildDirs returns services that declare build: and the host directory to pass
// as docker build context (absolute path). projectRoot is the compose project directory.
func ServiceBuildDirs(composeYAML []byte, projectRoot string) (map[string]string, error) {
	projectRoot = filepath.Clean(projectRoot)
	var root map[string]interface{}
	if err := yaml.Unmarshal(composeYAML, &root); err != nil {
		return nil, fmt.Errorf("parse compose yaml: %w", err)
	}
	if root == nil {
		return nil, fmt.Errorf("compose file is empty")
	}
	svcObj, ok := root["services"]
	if !ok || svcObj == nil {
		return nil, fmt.Errorf("no services: key in compose file")
	}
	services, ok := svcObj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("services must be a mapping")
	}
	out := make(map[string]string)
	for name, raw := range services {
		if raw == nil {
			continue
		}
		svc, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if !hasBuild(svc) {
			continue
		}
		dir, err := resolveBuildContextDir(projectRoot, svc["build"])
		if err != nil {
			return nil, fmt.Errorf("service %q: %w", name, err)
		}
		out[name] = dir
	}
	return out, nil
}

func resolveBuildContextDir(projectRoot string, buildVal interface{}) (string, error) {
	switch v := buildVal.(type) {
	case string:
		ctx := strings.TrimSpace(v)
		if ctx == "" {
			ctx = "."
		}
		return filepath.Abs(filepath.Join(projectRoot, filepath.FromSlash(ctx)))
	case map[string]interface{}:
		ctx := "."
		if s, ok := v["context"].(string); ok && strings.TrimSpace(s) != "" {
			ctx = strings.TrimSpace(s)
		}
		return filepath.Abs(filepath.Join(projectRoot, filepath.FromSlash(ctx)))
	default:
		return "", fmt.Errorf("unsupported build: type %T", buildVal)
	}
}
