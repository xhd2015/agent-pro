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
	// SnapshotSourceAgentRun is ttywatch.SnapshotText via an agent-run codex-tty.
	SnapshotSourceAgentRun = "agent-run"
)

// AgentRunSnapshotResult is a successful prefer-path capture (or dry-run resolve).
type AgentRunSnapshotResult struct {
	AgentRunSessionID string
	Contents          string
}

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

// tryAgentRunSnapshot looks up codexSessionID in the agent-run store and, when a
// live TTY registry entry exists, returns a VT-rendered snapshot (or dry-run hit).
func tryAgentRunSnapshot(agentRunHome, codexSessionID string, dryRun bool) (hit *AgentRunSnapshotResult, warn string, ok bool) {
	home, err := resolveAgentRunHome(agentRunHome)
	if err != nil || home == "" {
		return nil, "", false
	}
	store, err := agentstorage.NewFileStore(home)
	if err != nil {
		return nil, "", false
	}
	meta, err := agentstorage.FindByCodexSessionID(store, codexSessionID)
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
func preferAgentRunSnapshot(opts *SnapshotOpts, codexSessionID string) (hit *AgentRunSnapshotResult, warn string, ok bool) {
	if opts == nil || opts.ForceITerm {
		return nil, "", false
	}
	if opts.AgentRunSnapshot != nil {
		res, err := opts.AgentRunSnapshot(codexSessionID)
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
	if opts.Contents != nil || opts.ListITerm != nil {
		return nil, "", false
	}
	return tryAgentRunSnapshot(opts.AgentRunHome, codexSessionID, opts.DryRun)
}
