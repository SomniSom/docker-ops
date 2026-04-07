package sshexec

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func expandHome(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(p, "~/")), nil
	}
	return p, nil
}

// AuthMethods returns SSH auth methods: optional identity file (if readable), then ssh-agent.
func AuthMethods(identityPath string) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if identityPath != "" {
		path, err := expandHome(identityPath)
		if err != nil {
			return nil, err
		}
		if path != "" {
			if pemBytes, err := os.ReadFile(path); err == nil {
				if signer, err := ssh.ParsePrivateKey(pemBytes); err == nil {
					methods = append(methods, ssh.PublicKeys(signer))
				}
			}
		}
	}
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			ag := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(ag.Signers))
		}
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("%s", locale.T("ssh.err.no_credentials"))
	}
	return methods, nil
}
