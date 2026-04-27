package deploy

import (
	"testing"

	"github.com/SomniSom/docker-ops/internal/config"
)

// Exercises the build/push decisions in RunArtifacts (deploy_push vs --build vs save/load).
func TestArtifactsBuildAndPushMatrix(t *testing.T) {
	f, tr := false, true
	tests := []struct {
		name       string
		deployPush *bool
		cliBuild   bool
		image      string
		reg        *bool
		wantBuild  bool
		wantPush   bool
	}{
		{
			name: "save_load no flags", deployPush: nil, cliBuild: false, image: "myapp:1", reg: nil,
			wantBuild: false, wantPush: false,
		},
		{
			name: "save_load --build", deployPush: nil, cliBuild: true, image: "myapp:1", reg: nil,
			wantBuild: true, wantPush: false,
		},
		{
			name: "save_load deploy_push", deployPush: &tr, cliBuild: false, image: "myapp:1", reg: nil,
			wantBuild: true, wantPush: false,
		},
		{
			name: "registry no build", deployPush: &f, cliBuild: false, image: "reg.io/ns/a:1", reg: &tr,
			wantBuild: false, wantPush: false,
		},
		{
			name: "registry --build implies push", deployPush: &f, cliBuild: true, image: "reg.io/ns/a:1", reg: &tr,
			wantBuild: true, wantPush: true,
		},
		{
			name: "registry deploy_push", deployPush: &tr, cliBuild: false, image: "reg.io/ns/a:1", reg: &tr,
			wantBuild: true, wantPush: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				DeployImage:       tt.image,
				DeployPush:        tt.deployPush,
				DeployUseRegistry: tt.reg,
			}
			useSL := ArtifactsUseSaveLoad(cfg)
			needBuild := EffectiveDeployPush(cfg) || tt.cliBuild
			needPush := !useSL && (EffectiveDeployPush(cfg) || tt.cliBuild)
			if needBuild != tt.wantBuild {
				t.Fatalf("needBuild=%v want %v (useSL=%v)", needBuild, tt.wantBuild, useSL)
			}
			if needPush != tt.wantPush {
				t.Fatalf("needPush=%v want %v (useSL=%v)", needPush, tt.wantPush, useSL)
			}
		})
	}
}

func TestDeployBuildRemote(t *testing.T) {
	f, tr := false, true
	cfg := &config.Config{DeployPush: &tr, DeployBuildRemote: &tr}
	if !DeployBuildRemote(cfg, RunOpts{}) {
		t.Fatal("want true when deploy_push and deploy_build_remote")
	}
	if DeployBuildRemote(&config.Config{DeployBuildRemote: &f}, RunOpts{Build: true}) {
		t.Fatal("want false when deploy_build_remote false")
	}
	if !DeployBuildRemote(&config.Config{DeployBuildRemote: &tr}, RunOpts{Build: true}) {
		t.Fatal("want true with only --build and deploy_build_remote")
	}
}
