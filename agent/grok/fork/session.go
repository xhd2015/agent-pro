package fork

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/shell"
)

// SessionLaunch is the shared in-process grok --resume --fork-session plan
// used by grok-fork Mode B and `agent-pro grok session fork`.
type SessionLaunch struct {
	SessionID    string
	NewSessionID string
	Dir          string
	Bin          string
	Env          []string
}

// ResumeForkArgv returns argv for `grok --resume <id> --fork-session`
// (no binary name). newSessionID, when set, appends `--session-id`.
func ResumeForkArgv(sessionID, newSessionID string) []string {
	argv := []string{"--resume", sessionID, "--fork-session"}
	if sid := strings.TrimSpace(newSessionID); sid != "" {
		argv = append(argv, "--session-id", sid)
	}
	return argv
}

// ResolveGrokBin returns grokBin if set, otherwise lookPath("grok").
func ResolveGrokBin(grokBin string, lookPath func(file string) (string, error)) (string, error) {
	if strings.TrimSpace(grokBin) != "" {
		return grokBin, nil
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	bin, err := lookPath("grok")
	if err != nil {
		return "", fmt.Errorf("grok not found on PATH: %w", err)
	}
	return bin, nil
}

// Argv is ResumeForkArgv(SessionID, NewSessionID).
func (s SessionLaunch) Argv() []string {
	return ResumeForkArgv(s.SessionID, s.NewSessionID)
}

// CommandLine is the unquoted display form `<bin> --resume <id> --fork-session`.
func (s SessionLaunch) CommandLine() string {
	parts := append([]string{s.Bin}, s.Argv()...)
	return strings.Join(parts, " ")
}

// QuotedCommandLine quotes each token for an iTerm follow-up or dry-run.
func (s SessionLaunch) QuotedCommandLine() string {
	parts := append([]string{s.Bin}, s.Argv()...)
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		quoted = append(quoted, shell.ShellQuote(p))
	}
	return strings.Join(quoted, " ")
}

// Run invokes runForeground with the grok fork argv. runForeground must be
// non-nil (Main / Fork apply production defaults).
func (s SessionLaunch) Run(runForeground func(bin string, argv []string, dir string, env []string) error) error {
	if runForeground == nil {
		return fmt.Errorf("RunForeground is nil")
	}
	return runForeground(s.Bin, s.Argv(), s.Dir, s.Env)
}

// Fork launches grok --resume --fork-session in dir via opts.RunForeground.
func Fork(opts *Options, sessionID, newSessionID, dir string) error {
	if opts == nil {
		opts = &Options{}
	}
	applyDefaults(opts)
	bin, err := ResolveGrokBin(opts.GrokBin, opts.LookPath)
	if err != nil {
		return err
	}
	return SessionLaunch{
		SessionID:    sessionID,
		NewSessionID: newSessionID,
		Dir:          dir,
		Bin:          bin,
		Env:          opts.Env,
	}.Run(opts.RunForeground)
}
