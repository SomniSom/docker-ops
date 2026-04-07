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

// composeSession carries loaded project config and either a local compose.Runner
// (docker on the same machine) or a remote-only session (local is nil) that runs
// docker compose via SSH using cfg remote_ssh / remote_path.
type composeSession struct {
	cfg       *config.Config
	localRoot string
	local     *compose.Runner // nil when using remote compose over SSH
}

// newComposeSession loads docker-ops.yml for the resolved project root. If remote
// fields are set, it returns a session without a local Runner. Otherwise it checks
// docker and the Compose v2 plugin on PATH and builds a Runner with project file
// and project name from config.
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

// projectRoot returns the absolute path to the project directory (where docker-ops.yml lives).
func (s *composeSession) projectRoot() string { return s.localRoot }

// Run executes docker compose with the given arguments (subcommand and flags only;
// project -p/-f are added by compose.Runner or the remote layer). Uses a plain
// subprocess locally or a non-interactive SSH bash session remotely. Stdin/stdout/stderr
// are inherited; no signal forwarding beyond the OS default for the child process.
func (s *composeSession) Run(args ...string) error {
	if s.local != nil {
		return s.local.Run(args...)
	}
	return remote.RunDockerCompose(s.cfg, false, args...)
}

// RunTTY runs docker compose when the user needs terminal semantics: local runs use
// runLocalComposeTTY (optional raw stdin, SIGINT/SIGTERM forwarded to the whole compose
// process tree). Remote runs use an SSH PTY session (sshexec forwards signals to the server).
func (s *composeSession) RunTTY(args ...string) error {
	if s.local != nil {
		return s.runLocalComposeTTY(args, false)
	}
	return remote.RunDockerCompose(s.cfg, true, args...)
}

// RunExecTTY runs docker compose exec -it (or the remote equivalent). Locally it puts
// stdin in raw mode when appropriate so Ctrl+D and line editing behave like docker,
// and forwards interrupt/terminate to the compose process tree. Remotely it uses
// RunDockerComposeInteractive (raw stdin + SSH signal forwarding).
func (s *composeSession) RunExecTTY(args ...string) error {
	if s.local != nil {
		return s.runLocalComposeTTY(args, true)
	}
	return remote.RunDockerComposeInteractive(s.cfg, args...)
}

// runLocalComposeTTY starts docker compose as a child process with stderr teed for
// plugin error hints. If rawStdin is true and stdin is a terminal, stdin is switched
// to raw mode for the duration of the command (used for exec -it). When stdin is a
// terminal, SIGINT and SIGTERM are caught and delivered to the child’s process group
// (Unix) or the child process (Windows) so streaming commands like logs -f stop on Ctrl+C.
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

// cfgRef returns the session’s loaded config (compose file, project name, remote settings, etc.).
func (s *composeSession) cfgRef() *config.Config { return s.cfg }
