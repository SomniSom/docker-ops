//go:build unix

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

// prepareComposeTTYChild configures c before Start so the immediate child becomes the
// leader of a new process group (Setpgid). signalComposeProcessTree can then signal the
// whole docker compose tree with Kill(-pid, sig).
//
// The parent process (dq) must stay in the TTY foreground process group so keyboard
// SIGINT is delivered to dq and we can forward (or escalate) to the child. Do not move
// the foreground group to the child: then dq would not receive Ctrl+C and could not
// recover if docker ignores SIGINT.
func prepareComposeTTYChild(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setpgid = true
}

// signalComposeProcessTree sends sig to the child's process group (negative PID).
func signalComposeProcessTree(p *os.Process, sig syscall.Signal) {
	if p == nil || p.Pid <= 0 {
		return
	}
	_ = syscall.Kill(-p.Pid, sig)
}
