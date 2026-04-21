package deploy

import (
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/sshexec"
	"golang.org/x/crypto/ssh"
)

// RunRemoteReup runs config-check (if app_config) then docker compose build --pull, up -d, ps on the remote host.
func RunRemoteReup(client *ssh.Client, cfg *config.Config) error {
	return runRemoteComposeChain(client, cfg, "", false, [][]string{
		{"build", "--pull"},
		{"up", "-d"},
		{"ps"},
	})
}

// RunRemoteArtifactsFinish runs config-check, then pull+up or up only (save/load), with optional compose file override.
// exportDeployImage adds export DEPLOY_IMAGE=... when true (compose file still references ${DEPLOY_IMAGE}).
func RunRemoteArtifactsFinish(client *ssh.Client, cfg *config.Config, composeFileOverride string, skipPull bool, exportDeployImage bool) error {
	var steps [][]string
	if skipPull {
		steps = [][]string{{"up", "-d"}}
	} else {
		steps = [][]string{{"pull"}, {"up", "-d"}}
	}
	return runRemoteComposeChain(client, cfg, composeFileOverride, exportDeployImage, steps)
}

func runRemoteComposeChain(client *ssh.Client, cfg *config.Config, composeFileOverride string, exportDeployImage bool, steps [][]string) error {
	if client == nil || cfg == nil {
		return nil
	}
	rp := strings.TrimSpace(cfg.RemotePath)
	check := RemoteConfigCheckScript(cfg, rp)
	cf := strings.TrimSpace(composeFileOverride)
	if cf == "" {
		cf = cfg.ComposeFile
	}
	var parts []string
	parts = append(parts, "cd "+sshexec.ShellQuote(rp))
	if exportDeployImage {
		di := strings.TrimSpace(cfg.DeployImage)
		if di != "" {
			parts = append(parts, "export DEPLOY_IMAGE="+sshexec.ShellQuote(di))
		}
	}
	parts = append(parts, check)
	base := []string{"compose", "-p", cfg.ComposeProjectName, "-f", cf}
	for _, step := range steps {
		args := append(append([]string{}, base...), step...)
		parts = append(parts, "docker "+sshexec.QuoteArgs(args))
	}
	script := strings.Join(parts, " && ")
	return sshexec.RunBash(client, script, false)
}
