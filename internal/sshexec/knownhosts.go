package sshexec

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/SomniSom/docker-ops/internal/locale"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// AcceptNewHostKeyCallback behaves like OpenSSH StrictHostKeyChecking=accept-new:
// unknown keys are appended to knownHostsPath; mismatches are rejected.
func AcceptNewHostKeyCallback(knownHostsPath string) (ssh.HostKeyCallback, error) {
	if err := ensureFile(knownHostsPath, 0o600); err != nil {
		return nil, err
	}
	base, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", locale.Tf("ssh.err.kh_file", knownHostsPath), err)
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := base(hostname, remote, key)
		if err == nil {
			return nil
		}
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			return appendKnownHostLine(knownHostsPath, hostname, remote, key)
		}
		return fmt.Errorf("%w (%s)", err, locale.T("ssh.err.host_key_suffix"))
	}, nil
}

func ensureFile(path string, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, mode)
		if err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}

func appendKnownHostLine(path, hostname string, remote net.Addr, key ssh.PublicKey) error {
	addresses := []string{knownhosts.Normalize(remote.String())}
	if hostname != "" {
		_, port, err := net.SplitHostPort(remote.String())
		if err == nil {
			addresses = append(addresses, knownhosts.Normalize(net.JoinHostPort(hostname, port)))
		}
	}
	line := knownhosts.Line(addresses, key)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s\n", line); err != nil {
		return err
	}
	return nil
}
