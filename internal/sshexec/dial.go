package sshexec

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"

	"golang.org/x/crypto/ssh"
)

// Dial opens an SSH client using cfg (remote_ssh, ssh_identity) and accept-new known_hosts.
func Dial(cfg *config.Config) (*ssh.Client, error) {
	if cfg == nil || !cfg.RemoteConfigured() {
		return nil, fmt.Errorf("%s", locale.T("err.remote_not_configured"))
	}
	user, addr, err := ParseUserHost(cfg.RemoteSSH)
	if err != nil {
		return nil, err
	}
	methods, err := AuthMethods(cfg.SSHIdentity)
	if err != nil {
		return nil, err
	}
	home, _ := os.UserHomeDir()
	khPath := filepath.Join(home, ".ssh", "known_hosts")
	hk, err := AcceptNewHostKeyCallback(khPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", locale.T("ssh.err.known_hosts"), err)
	}
	sshCfg := &ssh.ClientConfig{
		User:            user,
		Auth:            methods,
		HostKeyCallback: hk,
		Timeout:         30 * time.Second,
	}
	client, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", locale.Tf("ssh.err.dial", addr), err)
	}
	return client, nil
}
