package agentrunapi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// GrokRunnerBindOpts configures EnsureGrokRunnerBound.
type GrokRunnerBindOpts struct {
	// GrokHome is the GROK_HOME / agent-runner-config-home root.
	// Empty falls back to agenttty.GrokHomeForRunner("").
	GrokHome string

	// DiscoverTimeout bounds on-disk discovery. Zero → 3s.
	DiscoverTimeout time.Duration

	// NotBefore lower-bounds session created_at during discovery.
	// Zero → derived from meta.CreatedAt.
	NotBefore time.Time
}

// EnsureGrokRunnerBound ensures meta.runner_session_id is bound for grok runners
// when a provider session id is discoverable. Late/fallback path when --open
// bind was missed so Classify can ModeResume after terminal close.
//
// Best-effort: bind misses do not return an error.
//
// Rules:
//  1. If meta.RunnerSessionID already set → no-op, bound=true (never overwrite)
//  2. Only grok / grok-tty runners
//  3. Prefer sessions/<id>/bind.json when state=ok with a non-empty id
//  4. Else discover via agenttty.DiscoverSession (workspace + InitialPrompt + NotBefore)
//  5. On success: store.UpdateSessionRunnerSessionID + return updated meta
func EnsureGrokRunnerBound(
	store agentstorage.Store,
	meta agentstorage.SessionMeta,
	opts GrokRunnerBindOpts,
) (updated agentstorage.SessionMeta, bound bool) {
	if id := strings.TrimSpace(meta.RunnerSessionID); id != "" {
		meta.RunnerSessionID = id
		return meta, true
	}
	if !isGrokRunner(meta.Runner) {
		return meta, false
	}

	sessionID := strings.TrimSpace(meta.SessionID)
	if sessionID == "" {
		return meta, false
	}

	// Prefer durable open-bind artifact written during --open.
	if id := readGrokBindJSONRunnerSessionID(store, sessionID); id != "" {
		return persistRunnerSessionID(store, meta, id)
	}

	grokHome := strings.TrimSpace(opts.GrokHome)
	if grokHome == "" {
		grokHome = agenttty.GrokHomeForRunner("")
	}
	if grokHome == "" {
		return meta, false
	}

	workspace := strings.TrimSpace(meta.Workspace)
	prompt := strings.TrimSpace(meta.InitialPrompt)
	if prompt == "" {
		// Without a prompt fingerprint, refuse to guess among sessions.
		return meta, false
	}

	notBefore := opts.NotBefore
	if notBefore.IsZero() {
		notBefore = parseMetaCreatedAt(meta.CreatedAt)
	}
	if notBefore.IsZero() {
		// DiscoverSession rejects sessions older than runStart-grace; without a
		// floor we could bind an unrelated older session.
		notBefore = time.Now().Add(-24 * time.Hour)
	}

	timeout := opts.DiscoverTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	id, _, err := agenttty.DiscoverSession(ctx, grokHome, workspace, prompt, notBefore)
	if err != nil || strings.TrimSpace(id) == "" {
		return meta, false
	}
	return persistRunnerSessionID(store, meta, strings.TrimSpace(id))
}

func isGrokRunner(runner string) bool {
	r := strings.ToLower(strings.TrimSpace(runner))
	return r == "grok-tty" || r == "grok" || strings.HasPrefix(r, "grok")
}

func productionGrokBindOpts(store agentstorage.Store, meta agentstorage.SessionMeta) GrokRunnerBindOpts {
	_ = store
	_ = meta
	return GrokRunnerBindOpts{
		GrokHome: agenttty.GrokHomeForRunner(strings.TrimSpace(meta.AgentRunnerConfigHome)),
	}
}

// readGrokBindJSONRunnerSessionID loads sessions/<id>/bind.json when state=ok.
func readGrokBindJSONRunnerSessionID(store agentstorage.Store, sessionID string) string {
	if store == nil {
		return ""
	}
	home := strings.TrimSpace(store.Home())
	sessionID = strings.TrimSpace(sessionID)
	if home == "" || sessionID == "" {
		return ""
	}
	path := filepath.Join(home, "sessions", sessionID, "bind.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var st struct {
		State           string `json:"state"`
		RunnerSessionID string `json:"runner_session_id"`
	}
	if json.Unmarshal(data, &st) != nil {
		return ""
	}
	if strings.ToLower(strings.TrimSpace(st.State)) != "ok" {
		return ""
	}
	return strings.TrimSpace(st.RunnerSessionID)
}
