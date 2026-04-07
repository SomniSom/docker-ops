//go:build !unix

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

// prepareComposeTTYChild is a no-op on non-Unix platforms; process groups are not used.
func prepareComposeTTYChild(c *exec.Cmd) {}

// signalComposeProcessTree sends sig only to the single process p (no process-group Kill).
func signalComposeProcessTree(p *os.Process, sig syscall.Signal) {
	if p == nil {
		return
	}
	_ = p.Signal(sig)
}
