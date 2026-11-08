package sessionactivity

import "sync"

// RunStatus answers whether one Session is executing a turn right now, as the
// runner's own agent registry sees it.
//
// known is false when the caller cannot answer for that Session: an unknown id,
// a runner without a registry, an unreachable daemon, or a Session started
// outside the daemon (a plain `dsh` TUI session, which nothing registers). A
// caller treats that as unknown rather than as idle, and does not probe
// session.lock. An answered true or false is authoritative.
type RunStatus func(sessionID string) (running bool, known bool)

var (
	runStatusMu sync.RWMutex
	runStatusFn RunStatus
)

// SetDefaultRunStatus installs the process-wide RunStatus that package-level
// Read and Follow consult for DSH Sessions; a nil hook uninstalls it and leaves
// DSH liveness unknown until a hook is installed again.
//
// Only the DSH reader consults it. Grok and Codex report liveness from their own
// registries, so installing a hook can never change their answer. Install it
// before reading: a hook decides which daemon answers, and a Session ID is
// answered only by the daemon that loaded it.
func SetDefaultRunStatus(fn RunStatus) {
	runStatusMu.Lock()
	runStatusFn = fn
	runStatusMu.Unlock()
}

// defaultRunStatus returns the installed hook, or nil when none is installed.
func defaultRunStatus() RunStatus {
	runStatusMu.RLock()
	defer runStatusMu.RUnlock()
	return runStatusFn
}

// runStatusHook is the hook this probe consults for the default DSH reader: its
// own field when set, else the installed default.
func (p Probe) runStatusHook() RunStatus {
	if p.RunStatus != nil {
		return p.RunStatus
	}
	return defaultRunStatus()
}

// pickLive is running vs not-running from the session/status hook only.
// A silent hook (nil, unknown id, denied, failed RPC) reports unknown; it
// never falls back to the process lease on session.lock.
func pickLive(sessionID string, rs RunStatus) (live, known bool) {
	if rs == nil {
		return false, false
	}
	return rs(sessionID)
}
