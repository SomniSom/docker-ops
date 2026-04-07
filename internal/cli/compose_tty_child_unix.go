//go:build unix

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

// prepareComposeTTYChild configures c before Start so the immediate child becomes the
// leader of a new process group (Setpgid). That allows signalComposeProcessTree to
// send sig to the entire tree (docker CLI plus the compose plugin subprocess), which
// plain Process.Signal(SIGINT) on the parent docker PID may not interrupt reliably
// during docker compose logs -f.
func prepareComposeTTYChild(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setpgid = true
}

// signalComposeProcessTree delivers sig to the process group whose ID equals p.Pid
// (Kill with a negative PID). No-op if p is nil or Pid is non-positive. Ignores errors
// from Kill (e.g. process already exited).
func signalComposeProcessTree(p *os.Process, sig syscall.Signal) {
	if p == nil || p.Pid <= 0 {
		return
	}
	_ = syscall.Kill(-p.Pid, sig)
}
