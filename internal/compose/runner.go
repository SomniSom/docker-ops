// Package compose runs docker compose as a subprocess (readme §4.4).
package compose

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SomniSom/docker-ops/internal/locale"
)

// Runner executes docker compose in projectRoot with project and file flags.
type Runner struct {
	ProjectRoot        string
	ComposeFile        string
	ComposeProjectName string
	DockerBin          string // default "docker"
}

func (r *Runner) docker() string {
	if r.DockerBin != "" {
		return r.DockerBin
	}
	return "docker"
}

// Command builds *exec.Cmd for docker compose with extra args (not run).
func (r *Runner) Command(args ...string) *exec.Cmd {
	full := append([]string{"compose", "-p", r.ComposeProjectName, "-f", r.ComposeFile}, args...)
	cmd := exec.Command(r.docker(), full...)
	cmd.Dir = r.ProjectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd
}

// RunCommand runs a command built with Command(), teeing stderr for compose-plugin hints on failure.
func (r *Runner) RunCommand(cmd *exec.Cmd) error {
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(cmd.Stderr, &stderrBuf)
	if err := cmd.Run(); err != nil {
		return HintIfComposePluginError(stderrBuf.String(), fmt.Errorf("%s: %w", locale.T("compose.run_prefix"), err))
	}
	return nil
}

// Run executes docker compose args and returns combined error from Wait.
func (r *Runner) Run(args ...string) error {
	composePath := filepath.Join(r.ProjectRoot, r.ComposeFile)
	if _, err := os.Stat(composePath); err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("compose.file_missing", composePath), err)
	}
	return r.RunCommand(r.Command(args...))
}

// LookPathDocker returns an error if docker is not found in PATH.
func LookPathDocker() error {
	_, err := exec.LookPath("docker")
	return err
}
