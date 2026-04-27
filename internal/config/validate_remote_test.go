package config

import "testing"

func TestValidateDeployBuildRemote_RequiresArtifacts(t *testing.T) {
	y := `
deploy_mode: source
deploy_build_remote: true
`
	if err := ValidateBytes("t.yaml", []byte(y)); err == nil {
		t.Fatalf("expected error for deploy_build_remote with source mode")
	}
}

func TestValidateDeployBuildRemote_ArtifactsOK(t *testing.T) {
	y := `
deploy_mode: artifacts
deploy_build_remote: true
`
	if err := ValidateBytes("t.yaml", []byte(y)); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
