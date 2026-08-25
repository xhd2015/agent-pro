package agentrunapi

import (
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// OpenSessionAction is the outcome kind of OpenSession.
const (
	OpenSessionActionFocused = "focused"
	OpenSessionActionResumed = "resumed"
)

// OpenSessionResult is returned by OpenSession.
type OpenSessionResult struct {
	Action    string // focused | resumed
	Runner    string
	SessionID string
}

// OpenSessionOpts drives OpenSession (agent-run session id keyed).
// Aligned with kck grok open's agent-run prefer path: live → focus;
// exited → ForceNew window running agent-run --auto-send-or-resume --open.
type OpenSessionOpts struct {
	Store     agentstorage.Store
	SessionID string
	// Driver is required for the ForceNew resume follow-up (e.g. marcus/spl +
	// ["agent-run"]). Empty Driver falls back to bare "agent-run" on PATH via
	// BuildFollowUpCommand / resolveFollowUpDriver.
	Driver agentdriver.Driver
	// AgentRunner overrides meta.Runner when building the follow-up (optional).
	AgentRunner string
	// WorkspaceDir overrides meta.Workspace for ForceNew cwd / --dir (optional).
	WorkspaceDir string

	Stderr io.Writer // unused today; reserved for soft warnings

	// Injectables (nil → production):
	FocusSession      func(FocusOpts) (FocusCandidate, error)
	OpenInNewTerminal func(OpenInNewTerminalOpts) error
}

// IsFocusMiss reports FocusSession failures that mean "no live host" so open
// should ForceNew a resume window. Includes missing registry JSON and empty
// iTerm match; does not treat multi-candidate ambiguity as a miss.
func IsFocusMiss(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "no iTerm candidates found") ||
		strings.Contains(msg, "registry entry not found") ||
		strings.Contains(msg, "registry pid missing")
}

// OpenSession focuses a live agent-run host tab, or ForceNews a new window that
// runs agent-run --auto-send-or-resume --open (does not resume in the caller's
// current terminal). SessionID is the agent-run / Marcus session id.
func OpenSession(opts OpenSessionOpts) (OpenSessionResult, error) {
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		return OpenSessionResult{}, fmt.Errorf("session id is required")
	}
	if opts.Store == nil {
		return OpenSessionResult{}, fmt.Errorf("store is required")
	}

	sess, err := opts.Store.GetSession(sessionID)
	runner := ""
	workspace := ""
	if err == nil && sess != nil {
		runner = strings.TrimSpace(sess.Meta.Runner)
		workspace = strings.TrimSpace(sess.Meta.Workspace)
	}
	if r := strings.TrimSpace(opts.AgentRunner); r != "" {
		runner = r
	}
	if w := strings.TrimSpace(opts.WorkspaceDir); w != "" {
		workspace = w
	}

	focusFn := opts.FocusSession
	if focusFn == nil {
		focusFn = FocusSession
	}
	_, ferr := focusFn(FocusOpts{
		Store:     opts.Store,
		SessionID: sessionID,
	})
	if ferr == nil {
		return OpenSessionResult{
			Action:    OpenSessionActionFocused,
			Runner:    runner,
			SessionID: sessionID,
		}, nil
	}
	if !IsFocusMiss(ferr) {
		return OpenSessionResult{}, ferr
	}

	if workspace == "" {
		return OpenSessionResult{}, fmt.Errorf("open resume requires workspace; session %s has empty workspace", sessionID)
	}

	openFn := opts.OpenInNewTerminal
	if openFn == nil {
		openFn = OpenInNewTerminal
	}
	if oerr := openFn(OpenInNewTerminalOpts{
		WorkspaceDir: workspace,
		FollowUpOpts: FollowUpOpts{
			Driver:                        opts.Driver,
			SessionID:                     sessionID,
			AgentRunner:                   runner,
			WorkspaceDir:                  workspace,
			Open:                          true,
			AllowRelocateResumeSessionDir: true,
		},
	}); oerr != nil {
		return OpenSessionResult{}, oerr
	}
	if runner == "" {
		runner = "(unknown)"
	}
	return OpenSessionResult{
		Action:    OpenSessionActionResumed,
		Runner:    runner,
		SessionID: sessionID,
	}, nil
}
