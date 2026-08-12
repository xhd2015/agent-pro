package agentrunapi

import (
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// CodexRunnerBindOpts configures EnsureCodexRunnerBound.
// All paths and scrollback are injectable — no process-global env/cwd required
// for product correctness under test.
type CodexRunnerBindOpts struct {
	// CodexHome is the CODEX_HOME root for rollout discovery.
	// Empty falls back to agenttty.CodexHome().
	CodexHome string

	// SnapshotScrollback, when non-nil, supplies TTY scrollback for resume-footer
	// bind without live registry/TCP. Nil means skip scrollback path.
	SnapshotScrollback func() (string, error)

	// NotBefore, when non-zero, is the lower bound for rollout session timestamps.
	// When zero, derived from meta.CreatedAt (RFC3339 / RFC3339Nano).
	NotBefore time.Time
}

// EnsureCodexRunnerBound ensures meta.runner_session_id is bound for codex runners
// when a Codex session id is discoverable. Discovery delegates to
// agenttty.DiscoverCodexSessionID (same recipe as open-time bind): cwd-matched
// rollout then optional scrollback resume footer. This is the late/fallback path
// that persists into the agent-run store when open-time bind was missed.
//
// Best-effort: does not return an error solely because bind missed; store update
// errors are soft (still return updated meta).
//
// Rules:
//  1. If meta.RunnerSessionID already set → no-op, bound=true (never overwrite)
//  2. Only codex runners (codex-tty / codex / prefix codex)
//  3. Discover via agenttty.DiscoverCodexSessionID (rollout + optional scrollback)
//  4. On success: store.UpdateSessionRunnerSessionID + return updated meta
func EnsureCodexRunnerBound(
	store agentstorage.Store,
	meta agentstorage.SessionMeta,
	opts CodexRunnerBindOpts,
) (updated agentstorage.SessionMeta, bound bool) {
	if id := strings.TrimSpace(meta.RunnerSessionID); id != "" {
		meta.RunnerSessionID = id
		return meta, true
	}
	if !isCodexRunner(meta.Runner) {
		return meta, false
	}

	codexHome := strings.TrimSpace(opts.CodexHome)
	if codexHome == "" {
		codexHome = agenttty.CodexHome()
	}
	workspace := strings.TrimSpace(meta.Workspace)
	if codexHome == "" {
		return meta, false
	}

	scrollback := ""
	if opts.SnapshotScrollback != nil {
		if text, err := opts.SnapshotScrollback(); err == nil {
			scrollback = text
		}
	}

	notBefore := opts.NotBefore
	if notBefore.IsZero() {
		notBefore = parseMetaCreatedAt(meta.CreatedAt)
	}

	// Same discovery as open-time bind (agenttty); we only add store persist.
	id, ok := agenttty.DiscoverCodexSessionID(codexHome, workspace, notBefore, scrollback)
	if !ok || strings.TrimSpace(id) == "" {
		return meta, false
	}
	return persistRunnerSessionID(store, meta, strings.TrimSpace(id))
}

func persistRunnerSessionID(store agentstorage.Store, meta agentstorage.SessionMeta, id string) (agentstorage.SessionMeta, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return meta, false
	}
	sessionID := strings.TrimSpace(meta.SessionID)
	if store != nil && sessionID != "" {
		_ = store.UpdateSessionRunnerSessionID(sessionID, id)
	}
	meta.RunnerSessionID = id
	return meta, true
}

func parseMetaCreatedAt(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t
	}
	return time.Time{}
}

// productionCodexBindOpts builds opts for live paths: default CodexHome and
// best-effort TTY scrollback snapshot when a terminal id is known.
func productionCodexBindOpts(store agentstorage.Store, meta agentstorage.SessionMeta) CodexRunnerBindOpts {
	opts := CodexRunnerBindOpts{
		CodexHome: agenttty.CodexHome(),
	}
	if store == nil {
		return opts
	}
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	if termID == "" {
		return opts
	}
	home := store.Home()
	opts.SnapshotScrollback = func() (string, error) {
		ttySess, err := agenttty.ResolveByTerminalID(home, termID)
		if err != nil || ttySess == nil || !ttySess.TCPReachable {
			return "", err
		}
		return ttywatch.SnapshotText(ttySess.Registry.ListenAddr, termID)
	}
	return opts
}
