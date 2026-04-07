// Package remote runs docker compose on a host over SSH (readme §4.5).
package remote

import (
	"errors"
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/sshexec"
	"golang.org/x/crypto/ssh"
)

// RunDockerCompose runs `docker compose` in cfg.RemotePath on the remote host.
func RunDockerCompose(cfg *config.Config, tty bool, composeArgs ...string) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return RunDockerComposeClient(client, cfg, "", tty, false, composeArgs...)
}

// RunDockerComposeInteractive runs docker compose over SSH for dq exec -it: raw local stdin + SIGINT forwarded to the server.
func RunDockerComposeInteractive(cfg *config.Config, composeArgs ...string) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return RunDockerComposeClient(client, cfg, "", true, true, composeArgs...)
}

// RunDockerComposeClient runs `docker compose` using an existing SSH client (same session as SFTP deploy).
// If composeFileOverride is non-empty, it is passed as -f instead of cfg.ComposeFile.
// rawLocalStdin should be true only for interactive exec (dq exec bash): Ctrl+D and signals behave correctly.
func RunDockerComposeClient(client *ssh.Client, cfg *config.Config, composeFileOverride string, tty, rawLocalStdin bool, composeArgs ...string) error {
	if client == nil || cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	rp := strings.TrimSpace(cfg.RemotePath)
	if rp == "" {
		return errors.New(locale.T("err.remote_ssh"))
	}
	cf := cfg.ComposeFile
	if strings.TrimSpace(composeFileOverride) != "" {
		cf = strings.TrimSpace(composeFileOverride)
	}
	inner := []string{"compose", "-p", cfg.ComposeProjectName, "-f", cf}
	inner = append(inner, composeArgs...)
	script := "cd " + sshexec.ShellQuote(rp)
	if di := strings.TrimSpace(cfg.DeployImage); di != "" {
		// Same as deploy.RunRemoteArtifactsFinish: compose files from gen-image-compose use ${DEPLOY_IMAGE}.
		script += " && export DEPLOY_IMAGE=" + sshexec.ShellQuote(di)
	}
	script += " && docker " + sshexec.QuoteArgs(inner)
	return sshexec.RunBashOpts(client, script, tty, sshexec.BashOpts{RawLocalStdin: rawLocalStdin})
}
