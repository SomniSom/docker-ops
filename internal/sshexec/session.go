package sshexec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/SomniSom/docker-ops/internal/locale"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// BashOpts tweaks interactive (TTY) SSH sessions (e.g. dq exec, dq logs -f).
type BashOpts struct {
	// RawLocalStdin puts the client's stdin in raw mode so keys (e.g. Ctrl+D / EOT) reach the remote shell.
	RawLocalStdin bool
	// CloseSessionOnInterrupt ends the SSH session on local Ctrl+C / SIGTERM instead of forwarding
	// ssh.SIGINT to the remote shell. Use for dq logs -f: dropping the session stops log streaming.
	// Do not set for dq exec -it, where SIGINT must reach the container shell.
	CloseSessionOnInterrupt bool
}

// RunBash runs bash -lc script on the SSH connection.
func RunBash(client *ssh.Client, script string, tty bool) error {
	return RunBashOpts(client, script, tty, BashOpts{})
}

// RunBashOpts is like RunBash with extra options for TTY sessions.
func RunBashOpts(client *ssh.Client, script string, tty bool, opts BashOpts) error {
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	fd := int(os.Stdin.Fd())
	if tty {
		sess.Stdin = os.Stdin
		if term.IsTerminal(fd) {
			w, h, err := term.GetSize(fd)
			if err == nil {
				modes := ssh.TerminalModes{
					ssh.ECHO:          1,
					ssh.TTY_OP_ISPEED: 14400,
					ssh.TTY_OP_OSPEED: 14400,
				}
				if err := sess.RequestPty("xterm-256color", h, w, modes); err != nil {
					return fmt.Errorf("%s: %w", locale.T("ssh.err.request_pty"), err)
				}
			}
		} else {
			_ = sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{})
		}
	}

	full := "set -e; " + script
	remote := "/bin/bash -lc " + ShellQuote(full)

	var oldTerm *term.State
	if tty && opts.RawLocalStdin && term.IsTerminal(fd) {
		var e error
		oldTerm, e = term.MakeRaw(fd)
		if e != nil {
			oldTerm = nil
		}
		if oldTerm != nil {
			defer func() { _ = term.Restore(fd, oldTerm) }()
		}
	}

	if !tty {
		if err := sess.Run(remote); err != nil {
			return fmt.Errorf("%s: %w", locale.T("ssh.err.remote"), err)
		}
		return nil
	}

	if err := sess.Start(remote); err != nil {
		return fmt.Errorf("%s: %w", locale.T("ssh.err.remote"), err)
	}

	sigCh := make(chan os.Signal, 8)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	var interruptCloseOnce sync.Once
	var closedSessionForInterrupt atomic.Bool
	go func() {
		for {
			select {
			case sig := <-sigCh:
				if opts.CloseSessionOnInterrupt {
					interruptCloseOnce.Do(func() {
						closedSessionForInterrupt.Store(true)
						_ = sess.Close()
					})
					continue
				}
				switch sig {
				case os.Interrupt:
					_ = sess.Signal(ssh.SIGINT)
				case syscall.SIGTERM:
					_ = sess.Signal(ssh.SIGTERM)
				}
			case <-done:
				return
			}
		}
	}()
	waitErr := sess.Wait()
	close(done)
	signal.Stop(sigCh)
	if closedSessionForInterrupt.Load() {
		return nil
	}
	if waitErr != nil {
		return fmt.Errorf("%s: %w", locale.T("ssh.err.remote"), waitErr)
	}
	return nil
}

// RunRemotePipe runs bash -lc script with stdin connected (e.g. docker save | ssh … docker load).
func RunRemotePipe(client *ssh.Client, script string, stdin io.Reader) error {
	if client == nil || stdin == nil {
		return fmt.Errorf("%s", locale.T("ssh.err.stdin"))
	}
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	sess.Stdin = stdin
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	full := "set -e; " + script
	remote := "/bin/bash -lc " + ShellQuote(full)
	if err := sess.Run(remote); err != nil {
		var ex *ssh.ExitError
		if errors.As(err, &ex) {
			return fmt.Errorf("%s: %w", locale.T("ssh.err.remote"), err)
		}
		return err
	}
	return nil
}
