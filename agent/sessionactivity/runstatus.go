package sessionactivity

import (
	"sync"
	"time"
)

// Lookup is one Session's answer from session/status.
//
// Answered is whether the daemon replied. Present is whether this id was in
// the map. Running is the map value when Present.
//
// The map only holds sessions the daemon has loaded, so absence is not
// evidence that a session is idle: a session whose daemon was restarted, or one
// running in a standalone TUI, is absent while its log may still hold an open
// turn. Absence therefore says only that no daemon process runs the turn, and
// the readers decide the phase from the folded log and the write lease
// instead. Present with Running false still reads the tail; not_found is
// reserved for a log the store cannot find at all, and an unanswered RPC is
// error.
type Lookup struct {
	Running  bool
	Present  bool
	Answered bool
	// At is when the answer was sampled. A cached hook reports the batch's
	// instant, so this can be older than the call: the readers compare it with an
	// open turn's start instead of assuming both reads are simultaneous.
	At time.Time
}

// RunStatus answers whether one Session is executing a turn right now, as the
// runner's own agent registry sees it. A silent hook (nil, failed RPC) leaves
// Answered false. An id missing from a successful map is Present false.
type RunStatus func(sessionID string) Lookup

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

func pickLookup(sessionID string, rs RunStatus) Lookup {
	if rs == nil {
		return Lookup{}
	}
	return rs(sessionID)
}
