package sessions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	// SnapshotSourceITerm is Contents / CaptureByTTY capture.
	SnapshotSourceITerm = "iterm"
	// SnapshotSourceAgentRun is ttywatch.SnapshotText via an agent-run grok-tty.
	SnapshotSourceAgentRun = "agent-run"
)

// AgentRunSnapshotResult is a successful prefer-path capture (or dry-run resolve).
type AgentRunSnapshotResult struct {
	AgentRunSessionID string
	Contents          string
}

// resolveAgentRunHome picks opts.AgentRunHome, else AGENT_RUN_HOME / ~/.agent-run.
func resolveAgentRunHome(explicit string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		return filepath.Clean(explicit), nil
	}
	if v := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME")); v != "" {
		return filepath.Clean(v), nil
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(dir, ".agent-run")), nil
}

// tryAgentRunSnapshot looks up grokSessionID in the agent-run store and, when a
// live TTY registry entry exists, returns a VT-rendered snapshot (or dry-run hit).
// Soft miss (not managed / unreachable / capture error) returns ok=false.
// Ambiguous mapping returns ok=false with a warning: prefix message for stderr.
func tryAgentRunSnapshot(agentRunHome, grokSessionID string, dryRun bool) (hit *AgentRunSnapshotResult, warn string, ok bool) {
	home, err := resolveAgentRunHome(agentRunHome)
	if err != nil || home == "" {
		return nil, "", false
	}
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		return nil, "", false
	}
	meta, err := agentstorage.FindByGrokSessionID(store, grokSessionID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "ambiguous") {
			return nil, "warning: " + msg + "; falling back to iTerm", false
		}
		return nil, "", false
	}
	arID := strings.TrimSpace(meta.SessionID)
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = arID
	}
	if termID == "" {
		return nil, "", false
	}

	entry, _, err := agenttty.LookupSession(store.Home(), termID)
	if err != nil {
		return nil, "", false
	}
	if dryRun {
		return &AgentRunSnapshotResult{AgentRunSessionID: arID}, "", true
	}
	text, err := ttywatch.SnapshotText(entry.ListenAddr, termID)
	if err != nil {
		return nil, "", false
	}
	return &AgentRunSnapshotResult{
		AgentRunSessionID: arID,
		Contents:          text,
	}, "", true
}

// preferAgentRunSnapshot runs the injectable hook or production lookup.
// Returns ok=false when the prefer path should be skipped.
func preferAgentRunSnapshot(opts *SnapshotOpts, grokSessionID string) (hit *AgentRunSnapshotResult, warn string, ok bool) {
	if opts == nil || opts.ForceITerm {
		return nil, "", false
	}
	if opts.AgentRunSnapshot != nil {
		res, err := opts.AgentRunSnapshot(grokSessionID)
		if err != nil {
			if strings.Contains(err.Error(), "ambiguous") {
				return nil, "warning: " + err.Error() + "; falling back to iTerm", false
			}
			return nil, "", false
		}
		if res == nil {
			return nil, "", false
		}
		return res, "", true
	}
	// Production only when no Contents/ListITerm inject (keeps L2 fakes off real home).
	if opts.Contents != nil || opts.ListITerm != nil {
		return nil, "", false
	}
	return tryAgentRunSnapshot(opts.AgentRunHome, grokSessionID, opts.DryRun)
}
