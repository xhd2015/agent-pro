package sessionactivity

import "sync"

// RunStatus answers whether one Session is executing a turn right now, as the
// runner's own agent registry sees it.
//
// known is false when the caller cannot answer for that Session: an unknown id,
// a runner without a registry, an unreachable daemon, or a Session started
// outside the daemon (a plain `dsh` TUI session, which nothing registers). A
// caller treats that as no opinion rather than as idle, and keeps the liveness
// it already reads. An answered true or false is authoritative: a Session whose
// lease says live while its Agent is idle is not running.
//
// Liveness is two facts with separate owners. Loaded — some process holds the
// Session lock, which is what the agent catalogue reports — stays the lease's
// answer. Running — an Agent is executing a turn — is this hook's answer.
type RunStatus func(sessionID string) (running bool, known bool)

var (
	runStatusMu sync.RWMutex
	runStatusFn RunStatus
)

// SetDefaultRunStatus installs the process-wide RunStatus that package-level
// Read and Follow consult for DSH Sessions; a nil hook uninstalls it and returns
// the DSH reader to lease-only liveness.
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

// pickLive resolves one Session's liveness from the process lease and the
// optional RunStatus hook.
//
// A hook that answers wins, because an Agent executing a turn is the fact a
// phase display needs: a lease held by a daemon that is not executing anything
// in that Session means loaded, not busy. When the hook has no answer — absent,
// unknown id, denied, failed RPC — the lease decides, so a runner without a
// registry keeps its previous behavior instead of reporting every Session idle.
func pickLive(sessionID string, leaseLive, leaseKnown bool, rs RunStatus) (live, known bool) {
	if rs != nil {
		if running, answered := rs(sessionID); answered {
			return running, true
		}
	}
	return leaseLive, leaseKnown
}
