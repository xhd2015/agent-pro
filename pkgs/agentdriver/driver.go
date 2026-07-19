// Package agentdriver resolves the host process argv used to re-exec into
// agent-run CLI (ForceNew follow-up and TTY __serve_* children).
//
// Implementation lives in github.com/xhd2015/tty-watch/pkgs/agentdriver so that
// agent-pro and standalone tty-watch share the same Driver type identity.
package agentdriver

import (
	"context"
	"os/exec"

	tw "github.com/xhd2015/tty-watch/pkgs/agentdriver"
)

// Driver is the host re-exec configuration for agent-run embedding.
// Type-aliased to tty-watch so values can be passed into ttywatch APIs.
type Driver = tw.Driver

// DefaultSelf returns Driver{Binary: abs path of this process, Args: nil}.
func DefaultSelf() (Driver, error) {
	return tw.DefaultSelf()
}

// Resolve returns a Driver with absolute Binary. Empty Binary uses DefaultSelf.
func Resolve(d Driver) (Driver, error) {
	return tw.Resolve(d)
}

// Command resolves d (if needed), builds argv with remainder, and returns *exec.Cmd.
func Command(d Driver, remainder ...string) (*exec.Cmd, error) {
	return tw.Command(d, remainder...)
}

// CommandContext is like Command with a context.
func CommandContext(ctx context.Context, d Driver, remainder ...string) (*exec.Cmd, error) {
	return tw.CommandContext(ctx, d, remainder...)
}

// MustArgv is Argv after Resolve; for tests and call sites that already handle Resolve.
func MustArgv(d Driver, remainder ...string) ([]string, error) {
	return tw.MustArgv(d, remainder...)
}
