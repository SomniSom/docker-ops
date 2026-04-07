package config

import (
	"os"
	"path/filepath"
)

const (
	dockerOpsYAML = "docker-ops.yaml"
	dockerOpsYML  = "docker-ops.yml"
	dqEnvFile     = "dq.env"
)

// ResolveProjectRoot returns the absolute project root from flag, DQ_PROJECT_ROOT, or cwd.
func ResolveProjectRoot(flagDir string) (string, error) {
	dir := flagDir
	if dir == "" {
		dir = os.Getenv("DQ_PROJECT_ROOT")
	}
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

// FindDockerOpsFile returns the path to docker-ops.yaml or docker-ops.yml in root, or "" if neither exists.
func FindDockerOpsFile(root string) string {
	for _, name := range []string{dockerOpsYAML, dockerOpsYML} {
		p := filepath.Join(root, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// DQEnvPath returns the path to dq.env in project root.
func DQEnvPath(root string) string {
	return filepath.Join(root, dqEnvFile)
}
