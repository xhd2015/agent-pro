package agentrunapi

import (
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// tryBindRunnerSessionFromZombie discovers a provider resume id from a keep-alive
// zombie TTY scrollback and persists it when meta.runner_session_id is empty.
//
// --open returns before codex discovery runs, so open sessions often exit
// unbound. After /exit the scrollback still holds
// "To continue this session, run codex resume <uuid>" — bind that so Classify
// can ModeResume instead of ModeRun.
//
// Best-effort: returns the (possibly updated) meta; never fails the call path.
func tryBindRunnerSessionFromZombie(store agentstorage.Store, meta agentstorage.SessionMeta) agentstorage.SessionMeta {
	if store == nil {
		return meta
	}
	if strings.TrimSpace(meta.RunnerSessionID) != "" {
		return meta
	}
	runner := strings.TrimSpace(meta.Runner)
	if !isCodexRunner(runner) {
		return meta
	}
	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	if termID == "" {
		return meta
	}
	ttySess, err := agenttty.ResolveByTerminalID(store.Home(), termID)
	if err != nil || ttySess == nil || !ttySess.TCPReachable {
		return meta
	}
	// Prefer durable exit signal; still allow footer-only when command_pid unknown.
	zombie := ttySess.Registry.CommandExited ||
		(ttySess.Registry.CommandPID > 0 && !ttywatch.ProcessAlive(ttySess.Registry.CommandPID))
	text, snapErr := ttywatch.SnapshotText(ttySess.Registry.ListenAddr, termID)
	if snapErr != nil || strings.TrimSpace(text) == "" {
		return meta
	}
	// Conservative: require exit footer AND (or Terminal exited) unless command_exited.
	if !zombie && !agenttty.CodexExitFooterPresent(text) && !agenttty.TerminalExitedMarkerPresent(text) {
		return meta
	}
	if zombie && !agenttty.CodexExitFooterPresent(text) && !agenttty.TerminalExitedMarkerPresent(text) {
		// command exited but no footer yet — still try extract resume id if present.
	}
	id := agenttty.FindCodexResumeSessionID(text)
	if id == "" {
		return meta
	}
	sessionID := strings.TrimSpace(meta.SessionID)
	if sessionID == "" {
		return meta
	}
	_ = store.UpdateSessionRunnerSessionID(sessionID, id)
	meta.RunnerSessionID = id
	return meta
}

func isCodexRunner(runner string) bool {
	r := strings.ToLower(strings.TrimSpace(runner))
	return r == "codex-tty" || r == "codex" || strings.HasPrefix(r, "codex")
}

// ReclaimZombieTerminalIDs force-frees TTY registry ids held by keep-alive
// serve after the agent child exited. Best-effort; safe to call when free.
func ReclaimZombieTerminalIDs(home, runner string, ids ...string) {
	home = strings.TrimSpace(home)
	if home == "" {
		return
	}
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		reclaimIDAcrossProviders(home, runner, id)
	}
}

func reclaimIDAcrossProviders(home, runner, sessionID string) {
	// Primary runner registry first.
	cfg := registryConfigForRunner(home, runner)
	if ttywatch.SessionIDInUse(cfg, sessionID) {
		_ = ttywatch.ReclaimSessionID(cfg, sessionID)
	}
	for _, p := range agenttty.ProviderListSorted() {
		subdir := p.RegistryDir
		if subdir == "" {
			continue
		}
		c := ttywatch.RegistryConfig{Home: home, Subdir: subdir}
		if ttywatch.SessionIDInUse(c, sessionID) {
			_ = ttywatch.ReclaimSessionID(c, sessionID)
		}
	}
}

func registryConfigForRunner(home, runner string) ttywatch.RegistryConfig {
	subdir := ""
	if p, ok := agenttty.Get(runner); ok {
		subdir = p.RegistryDir
	}
	if subdir == "" {
		r := strings.TrimSpace(runner)
		if r == "" {
			r = "grok-tty"
		}
		subdir = r + "-registry"
	}
	return ttywatch.RegistryConfig{Home: home, Subdir: subdir}
}
