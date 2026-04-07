//go:build unix

package cli

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// prepareComposeTTYChild configures c before Start so the immediate child becomes the
// leader of a new process group (Setpgid). Together with grantComposeTTYForeground and
// signalComposeProcessTree, this lets Ctrl+C stop docker compose logs -f reliably.
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

// grantComposeTTYForeground sets the controlling tty's foreground process group to the
// child's group (child PID equals PGID after Setpgid). Keyboard SIGINT is delivered to
// that group, so docker compose and its plugin receive the interrupt without relying
// solely on the parent forwarding signals.
//
// Returns the previous foreground PGRP and ok=true when the ioctl succeeded; callers
// must call restoreComposeTTYForeground on success before returning.
func grantComposeTTYForeground(ttyFD, childPID int) (savedPgrp int, ok bool) {
	if ttyFD < 0 || childPID <= 0 {
		return 0, false
	}
	saved, err := unix.IoctlGetInt(ttyFD, unix.TIOCGPGRP)
	if err != nil {
		return 0, false
	}
	if err := unix.IoctlSetPointerInt(ttyFD, unix.TIOCSPGRP, childPID); err != nil {
		return 0, false
	}
	return saved, true
}

// restoreComposeTTYForeground restores the tty foreground group saved by grantComposeTTYForeground.
func restoreComposeTTYForeground(ttyFD, savedPgrp int) {
	if ttyFD < 0 || savedPgrp <= 0 {
		return
	}
	_ = unix.IoctlSetPointerInt(ttyFD, unix.TIOCSPGRP, savedPgrp)
}
