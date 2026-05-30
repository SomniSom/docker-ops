package sshexec

import (
	"bytes"
	"fmt"

	"github.com/SomniSom/docker-ops/internal/locale"
	"golang.org/x/crypto/ssh"
)

// RunBashCapture runs bash -lc script and returns combined stdout (stderr goes to error on failure).
func RunBashCapture(client *ssh.Client, script string) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("%s", locale.T("err.remote_ssh"))
	}
	sess, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	var stdout bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stdout
	full := "set -e; " + script
	remote := "/bin/bash -lc " + ShellQuote(full)
	if err := sess.Run(remote); err != nil {
		return stdout.Bytes(), fmt.Errorf("%s: %w", locale.T("ssh.err.remote"), err)
	}
	return stdout.Bytes(), nil
}
