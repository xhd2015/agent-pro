package agentrunapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// FollowUpOpts builds a shell-quoted ForceNew child command (no --new-terminal).
type FollowUpOpts struct {
	// DriverBinary empty → "agent-run". Use os.Executable() for self re-exec.
	DriverBinary string
	// DriverArgsPrefix optional tokens after binary, before "run" (spl helper).
	DriverArgsPrefix []string

	SessionID                     string
	Prompt                        string
	AgentRunner                   string
	WorkspaceDir                  string
	NoSubmit                      bool
	AllowRelocateResumeSessionDir bool
	Open                          bool
	Detach                        bool
	// Env each "KEY=VALUE" → -e KEY=VALUE before -- / prompt (optional).
	Env []string
}

// BuildFollowUpCommand returns a single shell-quoted command line suitable for
// iterm2 FollowUpCommands. Never includes --new-terminal.
// Open/Detach mutual exclusion: if both true, return error.
// Empty SessionID → error.
func BuildFollowUpCommand(opts FollowUpOpts) (string, error) {
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		return "", fmt.Errorf("session id is required")
	}
	if opts.Open && opts.Detach {
		return "", fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}

	driver := strings.TrimSpace(opts.DriverBinary)
	if driver == "" {
		driver = "agent-run"
	}

	tokens := make([]string, 0, 16+len(opts.DriverArgsPrefix)+len(opts.Env))
	tokens = append(tokens, driver)
	for _, p := range opts.DriverArgsPrefix {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		tokens = append(tokens, p)
	}
	tokens = append(tokens, "run")
	tokens = append(tokens, "--session-id="+sessionID)
	if runner := strings.TrimSpace(opts.AgentRunner); runner != "" {
		tokens = append(tokens, "--agent-runner="+runner)
	}
	tokens = append(tokens, "--auto-send-or-resume")
	if dir := strings.TrimSpace(opts.WorkspaceDir); dir != "" {
		tokens = append(tokens, "--dir="+dir)
	}
	if opts.AllowRelocateResumeSessionDir {
		tokens = append(tokens, "--allow-relocate-resume-session-dir")
	}
	if opts.NoSubmit {
		tokens = append(tokens, "--no-submit")
	}
	if opts.Open {
		tokens = append(tokens, "--open")
	}
	if opts.Detach {
		tokens = append(tokens, "--detach")
	}
	for _, e := range opts.Env {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		tokens = append(tokens, "-e", e)
	}

	prompt := opts.Prompt
	// ForceNew auto-send child: open/detach always use "--"; non-empty prompt also
	// uses "--" so dash-leading prompts are not parsed as flags.
	if opts.Open || opts.Detach {
		tokens = append(tokens, "--", prompt)
	} else if strings.TrimSpace(prompt) != "" {
		tokens = append(tokens, "--", prompt)
	}

	quoted := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		quoted = append(quoted, shell.ShellQuote(tok))
	}
	return strings.Join(quoted, " "), nil
}

// OpenInNewTerminalOpts opens a new terminal with a follow-up command.
type OpenInNewTerminalOpts struct {
	WorkspaceDir string
	// FollowUp if non-empty is used as-is; else built from FollowUpOpts.
	FollowUp     string
	FollowUpOpts FollowUpOpts
	// OpenTerminal replaces production iTerm ForceNew when set (unit tests).
	OpenTerminal func(dir string, followUp string) error
}

// OpenInNewTerminal invokes OpenTerminal(dir, followUp). When OpenTerminal is
// nil, production uses iterm2 ModeForceNew.
func OpenInNewTerminal(opts OpenInNewTerminalOpts) error {
	dir := strings.TrimSpace(opts.WorkspaceDir)
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("workspace dir: %w", err)
		}
		dir = cwd
	} else {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("workspace dir: %w", err)
		}
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			abs = resolved
		}
		dir = abs
	}

	followUp := strings.TrimSpace(opts.FollowUp)
	if followUp == "" {
		var err error
		// Prefer FollowUpOpts.WorkspaceDir from opts if unset so built command has --dir.
		fuOpts := opts.FollowUpOpts
		if strings.TrimSpace(fuOpts.WorkspaceDir) == "" && strings.TrimSpace(opts.WorkspaceDir) != "" {
			fuOpts.WorkspaceDir = opts.WorkspaceDir
		}
		followUp, err = BuildFollowUpCommand(fuOpts)
		if err != nil {
			return err
		}
	}

	open := opts.OpenTerminal
	if open == nil {
		open = defaultOpenTerminal
	}
	return open(dir, followUp)
}

func defaultOpenTerminal(dir, followUp string) error {
	return iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
	})
}
