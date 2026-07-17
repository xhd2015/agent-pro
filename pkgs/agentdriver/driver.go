// Package agentdriver resolves the host process argv used to re-exec into
// agent-run CLI (ForceNew follow-up and TTY __serve_* children).
//
// Callers pass an optional Driver; empty Binary applies DefaultSelf once.
// Embedding hosts (e.g. spl) set Binary + Args explicitly:
//
//	Driver{Binary: abs(spl), Args: []string{"agent-run"}}
//
// Runtime argv is always:
//
//	[abs(Binary), Args..., remainder...]
package agentdriver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Driver is the host re-exec configuration for agent-run embedding.
type Driver struct {
	// Binary is the executable path or bare name. Empty → DefaultSelf at Resolve.
	Binary string
	// Args are tokens after Binary and before remainder (e.g. "agent-run" for spl).
	Args []string
}

// DefaultSelf returns Driver{Binary: abs path of this process, Args: nil}.
func DefaultSelf() (Driver, error) {
	exe, err := selfExecutable()
	if err != nil {
		return Driver{}, err
	}
	return Driver{Binary: exe}, nil
}

// Resolve returns a Driver with absolute Binary. Empty Binary uses DefaultSelf.
// Empty Args entries are dropped. Args from DefaultSelf stay nil when input Args empty.
func Resolve(d Driver) (Driver, error) {
	bin := strings.TrimSpace(d.Binary)
	args := cleanArgs(d.Args)
	if bin == "" {
		self, err := DefaultSelf()
		if err != nil {
			return Driver{}, err
		}
		// Explicit Args on zero Binary still apply after defaulting the binary.
		if len(args) > 0 {
			self.Args = args
		}
		return self, nil
	}
	abs, err := absExecutable(bin)
	if err != nil {
		return Driver{}, err
	}
	return Driver{Binary: abs, Args: args}, nil
}

// Argv builds [Binary, Args..., remainder...]. Caller should Resolve first so
// Binary is absolute; Argv does not re-default empty Binary (returns error).
func (d Driver) Argv(remainder ...string) ([]string, error) {
	bin := strings.TrimSpace(d.Binary)
	if bin == "" {
		return nil, fmt.Errorf("agentdriver: empty Binary (call Resolve first)")
	}
	n := 1 + len(d.Args) + len(remainder)
	out := make([]string, 0, n)
	out = append(out, bin)
	out = append(out, cleanArgs(d.Args)...)
	for _, r := range remainder {
		out = append(out, r)
	}
	return out, nil
}

// Command resolves d (if needed), builds argv with remainder, and returns *exec.Cmd.
func Command(d Driver, remainder ...string) (*exec.Cmd, error) {
	return CommandContext(context.Background(), d, remainder...)
}

// CommandContext is like Command with a context.
func CommandContext(ctx context.Context, d Driver, remainder ...string) (*exec.Cmd, error) {
	resolved, err := Resolve(d)
	if err != nil {
		return nil, err
	}
	argv, err := resolved.Argv(remainder...)
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, argv[0], argv[1:]...), nil
}

// MustArgv is Argv after Resolve; for tests and call sites that already handle Resolve.
func MustArgv(d Driver, remainder ...string) ([]string, error) {
	resolved, err := Resolve(d)
	if err != nil {
		return nil, err
	}
	return resolved.Argv(remainder...)
}

func cleanArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, 0, len(args))
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		out = append(out, a)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func selfExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		if len(os.Args) > 0 && strings.TrimSpace(os.Args[0]) != "" {
			exe = os.Args[0]
		} else {
			return "", fmt.Errorf("agentdriver: resolve self executable: %w", err)
		}
	}
	return absExecutable(exe)
}

func absExecutable(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("agentdriver: empty binary path")
	}
	// Bare names without path separators: keep for PATH lookup by exec; still Abs if possible.
	if !strings.Contains(path, string(filepath.Separator)) && path != "." {
		if abs, err := exec.LookPath(path); err == nil {
			path = abs
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("agentdriver: abs %q: %w", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return abs, nil
}
