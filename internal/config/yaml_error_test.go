package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFormatYAMLParseError_MisindentedDeployInclude(t *testing.T) {
	yml := `# header
remote_ssh: u@h
remote_path: /p
compose_project_name: x
 deploy_include:
   - a
`
	var cfg Config
	parseErr := yaml.Unmarshal([]byte(yml), &cfg)
	if parseErr == nil {
		t.Fatal("expected yaml parse error")
	}
	err := FormatYAMLParseError("docker-ops.yml", []byte(yml), parseErr)
	out := err.Error()
	if !strings.Contains(out, "deploy_include") {
		t.Fatalf("expected key name in message:\n%s", out)
	}
	if !strings.Contains(out, "Context") {
		t.Fatalf("expected context block:\n%s", out)
	}
}

func TestScanMisindentedRootKeys(t *testing.T) {
	lines := strings.Split(`a: 1
 deploy_include:
  - x`, "\n")
	s := scanMisindentedRootKeys(lines)
	if !strings.Contains(s, "deploy_include") || !strings.Contains(s, "line 2") {
		t.Fatalf("got %q", s)
	}
}
