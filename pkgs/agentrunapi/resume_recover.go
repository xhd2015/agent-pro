package agentrunapi

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// localGrokRunnerSessionMissing reports whether a grok/grok-tty ModeResume
// should be skipped because meta.runner_session_id has no on-disk updates.jsonl
// under GROK_HOME. Non-grok runners always return false (no precheck).
//
// When true, callers clear the stale bind and ModeRun instead of launching
// `grok --resume <id>` (which would fail with remote restore 404).
func localGrokRunnerSessionMissing(opts Opts, meta agentstorage.SessionMeta) bool {
	runner := effectiveRunner(opts, meta)
	if !isGrokRunnerName(runner) {
		return false
	}
	id := strings.TrimSpace(meta.RunnerSessionID)
	if id == "" {
		return true
	}
	configHome := strings.TrimSpace(opts.RunnerConfigHome)
	if configHome == "" {
		configHome = strings.TrimSpace(meta.AgentRunnerConfigHome)
	}
	grokHome := agenttty.GrokHomeForRunner(configHome)
	workspace := strings.TrimSpace(opts.WorkspaceDir)
	if workspace == "" {
		workspace = strings.TrimSpace(meta.Workspace)
	}
	_, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, id)
	return !ok
}

func isGrokRunnerName(runner string) bool {
	r := strings.ToLower(strings.TrimSpace(runner))
	return r == "grok-tty" || r == "grok" || strings.HasPrefix(r, "grok")
}

// clearStaleRunnerSessionBind drops meta.runner_session_id and bind.json so
// Classify / ModeRun can start fresh. Best-effort on bind.json removal.
func clearStaleRunnerSessionBind(store agentstorage.Store, sessionID, staleRunnerSessionID string, stderr io.Writer) error {
	if store == nil {
		return fmt.Errorf("store is required")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if err := store.ClearSessionRunnerSessionID(sessionID); err != nil {
		return err
	}
	bindPath := filepath.Join(store.Home(), "sessions", sessionID, "bind.json")
	_ = os.Remove(bindPath)
	if stderr == nil {
		stderr = os.Stderr
	}
	stale := strings.TrimSpace(staleRunnerSessionID)
	if stale != "" {
		fmt.Fprintf(stderr, "warning: bound runner session %s has no local grok data; clearing bind and starting a fresh session\n", stale)
	} else {
		fmt.Fprintf(stderr, "warning: bound runner session has no local grok data; clearing bind and starting a fresh session\n")
	}
	return nil
}
