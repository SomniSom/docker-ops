package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/SomniSom/docker-ops/internal/compose"
	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/remote"
	"golang.org/x/term"
)

// composeSession runs docker compose locally or over SSH (readme §4.5).
type composeSession struct {
	cfg       *config.Config
	localRoot string
	local     *compose.Runner
}

func newComposeSession(projectDir *string) (*composeSession, error) {
	cfg, root, err := loadCfg(projectDir)
	if err != nil {
		return nil, err
	}
	if cfg.RemoteConfigured() {
		return &composeSession{cfg: cfg, localRoot: root}, nil
	}
	if err := compose.LookPathDocker(); err != nil {
		return nil, err
	}
	if err := compose.RequireComposeV2Plugin(); err != nil {
		return nil, err
	}
	r := &compose.Runner{
		ProjectRoot:        root,
		ComposeFile:        cfg.ComposeFile,
		ComposeProjectName: cfg.ComposeProjectName,
	}
	return &composeSession{cfg: cfg, localRoot: root, local: r}, nil
}

func (s *composeSession) projectRoot() string { return s.localRoot }

func (s *composeSession) Run(args ...string) error {
	if s.local != nil {
		return s.local.Run(args...)
	}
	return remote.RunDockerCompose(s.cfg, false, args...)
}

func (s *composeSession) RunTTY(args ...string) error {
	if s.local != nil {
		return s.runLocalComposeTTY(args, false)
	}
	return remote.RunDockerCompose(s.cfg, true, args...)
}

// RunExecTTY runs interactive docker compose exec (-it): raw stdin locally, SIGINT to child / SSH session.
func (s *composeSession) RunExecTTY(args ...string) error {
	if s.local != nil {
		return s.runLocalComposeTTY(args, true)
	}
	return remote.RunDockerComposeInteractive(s.cfg, args...)
}

func (s *composeSession) runLocalComposeTTY(args []string, rawStdin bool) error {
	if s.local == nil {
		return fmt.Errorf("dq: internal error: runLocalComposeTTY without local runner")
	}
	composePath := filepath.Join(s.local.ProjectRoot, s.local.ComposeFile)
	if _, err := os.Stat(composePath); err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("compose.file_missing", composePath), err)
	}
	c := s.local.Command(args...)
	prepareComposeTTYChild(c)
	var stderrBuf bytes.Buffer
	c.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	fd := int(os.Stdin.Fd())
	var oldState *term.State
	if rawStdin && term.IsTerminal(fd) {
		var err error
		oldState, err = term.MakeRaw(fd)
		if err != nil {
			oldState = nil
		}
		if oldState != nil {
			defer func() { _ = term.Restore(fd, oldState) }()
		}
	}

	if err := c.Start(); err != nil {
		return compose.HintIfComposePluginError(stderrBuf.String(), fmt.Errorf("%s: %w", locale.T("compose.run_prefix"), err))
	}

	if term.IsTerminal(fd) {
		savedPgrp, fgOK := grantComposeTTYForeground(fd, c.Process.Pid)
		if fgOK {
			defer restoreComposeTTYForeground(fd, savedPgrp)
		}

		sigCh := make(chan os.Signal, 8)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		done := make(chan struct{})
		go func() {
			for {
				select {
				case sig := <-sigCh:
					switch sig {
					case os.Interrupt:
						signalComposeProcessTree(c.Process, syscall.SIGINT)
					case syscall.SIGTERM:
						signalComposeProcessTree(c.Process, syscall.SIGTERM)
					}
				case <-done:
					return
				}
			}
		}()
		err := c.Wait()
		close(done)
		signal.Stop(sigCh)
		if err != nil {
			return compose.HintIfComposePluginError(stderrBuf.String(), fmt.Errorf("%s: %w", locale.T("compose.run_prefix"), err))
		}
		return nil
	}

	if err := c.Wait(); err != nil {
		return compose.HintIfComposePluginError(stderrBuf.String(), fmt.Errorf("%s: %w", locale.T("compose.run_prefix"), err))
	}
	return nil
}

func (s *composeSession) cfgRef() *config.Config { return s.cfg }
