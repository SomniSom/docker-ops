//go:build !unix

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

func prepareComposeTTYChild(c *exec.Cmd) {}

func signalComposeProcessTree(p *os.Process, sig syscall.Signal) {
	if p == nil {
		return
	}
	_ = p.Signal(sig)
}

func grantComposeTTYForeground(ttyFD, childPID int) (savedPgrp int, ok bool) {
	return 0, false
}

func restoreComposeTTYForeground(ttyFD, savedPgrp int) {}
