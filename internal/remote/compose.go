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
func RunDockerCompose(cfg *config.Config, projectRoot string, tty bool, composeArgs ...string) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return RunDockerComposeClient(client, cfg, projectRoot, "", tty, false, composeArgs...)
}

// RunDockerComposeInteractive runs docker compose over SSH for dq exec -it: raw local stdin + SIGINT forwarded to the server.
func RunDockerComposeInteractive(cfg *config.Config, projectRoot string, composeArgs ...string) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return RunDockerComposeClient(client, cfg, projectRoot, "", true, true, composeArgs...)
}

// RunDockerComposeClient runs `docker compose` using an existing SSH client (same session as SFTP deploy).
func RunDockerComposeClient(client *ssh.Client, cfg *config.Config, projectRoot, composeFileOverride string, tty, rawLocalStdin bool, composeArgs ...string) error {
	if client == nil || cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	script, err := dockerComposeScript(cfg, projectRoot, composeFileOverride, composeArgs...)
	if err != nil {
		return err
	}
	return sshexec.RunBashOpts(client, script, tty, sshexec.BashOpts{
		RawLocalStdin:           rawLocalStdin,
		CloseSessionOnInterrupt: tty && !rawLocalStdin,
	})
}

// DockerComposeOutput runs docker compose on the remote host and returns stdout.
func DockerComposeOutput(cfg *config.Config, projectRoot string, composeArgs ...string) ([]byte, error) {
	if cfg == nil || !cfg.RemoteConfigured() {
		return nil, errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	script, err := dockerComposeScript(cfg, projectRoot, "", composeArgs...)
	if err != nil {
		return nil, err
	}
	return sshexec.RunBashCapture(client, script)
}

// RunDocker runs `docker` on the remote host (stdout/stderr inherited).
func RunDocker(cfg *config.Config, dockerArgs ...string) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return sshexec.RunBash(client, dockerScript(cfg, dockerArgs...), false)
}

// DockerOutput runs docker on the remote host and returns stdout.
func DockerOutput(cfg *config.Config, dockerArgs ...string) ([]byte, error) {
	if cfg == nil || !cfg.RemoteConfigured() {
		return nil, errors.New(locale.T("err.remote_ssh"))
	}
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return sshexec.RunBashCapture(client, dockerScript(cfg, dockerArgs...))
}

func dockerScript(cfg *config.Config, dockerArgs ...string) string {
	script := "docker " + sshexec.QuoteArgs(dockerArgs)
	if cfg != nil {
		if rp := strings.TrimSpace(cfg.RemotePath); cfg.RemoteConfigured() && rp != "" {
			script = "cd " + sshexec.ShellQuote(rp) + " && " + script
		}
	}
	return script
}

func dockerComposeScript(cfg *config.Config, projectRoot, composeFileOverride string, composeArgs ...string) (string, error) {
	if cfg == nil || !cfg.RemoteConfigured() {
		return "", errors.New(locale.T("err.remote_ssh"))
	}
	rp := strings.TrimSpace(cfg.RemotePath)
	if rp == "" {
		return "", errors.New(locale.T("err.remote_ssh"))
	}
	cf := config.EffectiveComposeFile(cfg, projectRoot)
	if strings.TrimSpace(composeFileOverride) != "" {
		cf = strings.TrimSpace(composeFileOverride)
	}
	inner := []string{"compose", "-p", cfg.ComposeProjectName, "-f", cf}
	inner = append(inner, composeArgs...)
	script := "cd " + sshexec.ShellQuote(rp)
	if di := strings.TrimSpace(cfg.DeployImage); di != "" {
		script += " && export DEPLOY_IMAGE=" + sshexec.ShellQuote(di)
	}
	script += " && docker " + sshexec.QuoteArgs(inner)
	return script, nil
}
