package agentrunapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// PromptFileSpillMinRunes is the exclusive lower bound for auto-spill:
// rune count of TrimSpace(Prompt) must be > this value to spill to --prompt-file.
// Locked: 600.
const PromptFileSpillMinRunes = 600

// FollowUpOpts builds a shell-quoted ForceNew child command (no --new-terminal).
type FollowUpOpts struct {
	// Driver is the host re-exec config. Zero Binary → Resolve defaults to self
	// unless legacy DriverBinary/DriverArgsPrefix are set (compat).
	Driver agentdriver.Driver
	// DriverBinary is deprecated: use Driver.Binary.
	// Empty with empty Driver → historical default "agent-run" for bare CLI PATH.
	DriverBinary string
	// DriverArgsPrefix is deprecated: use Driver.Args.
	DriverArgsPrefix []string

	SessionID                     string
	Prompt                        string
	AgentRunner                   string
	WorkspaceDir                  string
	NoSubmit                      bool
	AllowRelocateResumeSessionDir bool
	Open                          bool
	Detach                        bool
	// Color true → emit --color after open/detach flags, before -e / -- / prompt.
	Color bool
	// Model is optional; when non-empty, BuildFollowUpCommand emits --model=<m>
	// (or two-token form). Empty → omit (no product default).
	Model string
	// ModelReasoningEffort is optional; when non-empty, emits
	// --model-reasoning-effort=<level>. Empty → omit (no default max).
	ModelReasoningEffort string
	// PromptFile when non-empty: emit --prompt-file=<abs>; omit Prompt body after --.
	// Caller already spilled; do not re-write the file.
	PromptFile string
	// PromptSpillDir when auto-spilling long Prompt: write under this dir (tests inject).
	// Empty production: filepath.Join(os.TempDir(), "agent-run-prompt-spill").
	PromptSpillDir string
	// Env each "KEY=VALUE" → -e KEY=VALUE before -- / prompt (optional).
	Env []string
	// ExtraArgs are inserted after -e flags and before -- / prompt (e.g. event-bus flags).
	ExtraArgs []string
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

	driver, err := resolveFollowUpDriver(opts)
	if err != nil {
		return "", err
	}

	remainder := make([]string, 0, 16+len(opts.Env))
	remainder = append(remainder, "run")
	remainder = append(remainder, "--session-id="+sessionID)
	if runner := strings.TrimSpace(opts.AgentRunner); runner != "" {
		remainder = append(remainder, "--agent-runner="+runner)
	}
	remainder = append(remainder, "--auto-send-or-resume")
	if dir := strings.TrimSpace(opts.WorkspaceDir); dir != "" {
		remainder = append(remainder, "--dir="+dir)
	}
	if opts.AllowRelocateResumeSessionDir {
		remainder = append(remainder, "--allow-relocate-resume-session-dir")
	}
	if opts.NoSubmit {
		remainder = append(remainder, "--no-submit")
	}
	if opts.Open {
		remainder = append(remainder, "--open")
	}
	if opts.Detach {
		remainder = append(remainder, "--detach")
	}
	if opts.Color {
		remainder = append(remainder, "--color")
	}
	if m := strings.TrimSpace(opts.Model); m != "" {
		remainder = append(remainder, "--model="+m)
	}
	if e := strings.TrimSpace(opts.ModelReasoningEffort); e != "" {
		remainder = append(remainder, "--model-reasoning-effort="+e)
	}
	for _, e := range opts.Env {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		remainder = append(remainder, "-e", e)
	}
	for _, a := range opts.ExtraArgs {
		if a == "" {
			continue
		}
		remainder = append(remainder, a)
	}

	// Prompt delivery priority:
	// 1. Explicit PromptFile → --prompt-file=abs; no argv body
	// 2. Long Prompt (> PromptFileSpillMinRunes after TrimSpace) → auto-spill + --prompt-file
	// 3. Else existing Open/Detach/`--` + prompt behavior
	if pf := strings.TrimSpace(opts.PromptFile); pf != "" {
		abs, err := filepath.Abs(pf)
		if err != nil {
			return "", fmt.Errorf("prompt-file path: %w", err)
		}
		remainder = append(remainder, "--prompt-file="+abs)
	} else if utf8.RuneCountInString(strings.TrimSpace(opts.Prompt)) > PromptFileSpillMinRunes {
		abs, err := spillFollowUpPrompt(opts)
		if err != nil {
			return "", err
		}
		remainder = append(remainder, "--prompt-file="+abs)
	} else {
		prompt := opts.Prompt
		// ForceNew auto-send child: open/detach always use "--"; non-empty prompt also
		// uses "--" so dash-leading prompts are not parsed as flags.
		if opts.Open || opts.Detach {
			remainder = append(remainder, "--", prompt)
		} else if strings.TrimSpace(prompt) != "" {
			remainder = append(remainder, "--", prompt)
		}
	}

	tokens, err := driver.Argv(remainder...)
	if err != nil {
		return "", err
	}

	quoted := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		quoted = append(quoted, shell.ShellQuote(tok))
	}
	return strings.Join(quoted, " "), nil
}

// spillFollowUpPrompt writes TrimSpace(Prompt) under PromptSpillDir (or a
// production temp default) and returns an absolute path for --prompt-file.
func spillFollowUpPrompt(opts FollowUpOpts) (string, error) {
	dir := strings.TrimSpace(opts.PromptSpillDir)
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "agent-run-prompt-spill")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("prompt spill dir: %w", err)
	}
	pattern := "prompt-*.txt"
	if sid := strings.TrimSpace(opts.SessionID); sid != "" {
		// Include session id for uniqueness / debugging (safe IDs only).
		safe := strings.Map(func(r rune) rune {
			if r == '/' || r == '\\' || r == 0 {
				return '_'
			}
			return r
		}, sid)
		pattern = "prompt-" + safe + "-*.txt"
	}
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("create prompt spill file: %w", err)
	}
	path := f.Name()
	body := strings.TrimSpace(opts.Prompt)
	if _, err := f.WriteString(body); err != nil {
		f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write prompt spill file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close prompt spill file: %w", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

// resolveFollowUpDriver merges Driver with legacy fields and Resolves.
// Historical empty DriverBinary defaulted to bare "agent-run" (PATH). When
// Driver is fully zero and legacy fields empty, keep that PATH name without
// forcing abs(self), so existing slack-msg / bare defaults stay valid.
func resolveFollowUpDriver(opts FollowUpOpts) (agentdriver.Driver, error) {
	d := opts.Driver
	if strings.TrimSpace(d.Binary) == "" {
		if b := strings.TrimSpace(opts.DriverBinary); b != "" {
			d.Binary = b
		}
	}
	if len(d.Args) == 0 && len(opts.DriverArgsPrefix) > 0 {
		d.Args = append([]string(nil), opts.DriverArgsPrefix...)
	}
	// Fully unspecified → historical "agent-run" on PATH (not DefaultSelf).
	if strings.TrimSpace(d.Binary) == "" && len(d.Args) == 0 {
		d.Binary = "agent-run"
		// Do not abs-resolve bare "agent-run" via DefaultSelf; LookPath in Resolve.
		return agentdriver.Resolve(d)
	}
	return agentdriver.Resolve(d)
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
