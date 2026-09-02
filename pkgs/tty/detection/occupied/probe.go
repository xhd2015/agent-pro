package occupied

import "time"

// Status is composer occupancy from a space probe.
type Status string

const (
	Empty    Status = "empty"
	Occupied Status = "occupied"
	Unknown  Status = "unknown"
)

func (s Status) String() string { return string(s) }

const defaultProbeSettle = 150 * time.Millisecond

// defaultStabilizeAttempts is how many post-DEL Snapshot polls Probe uses to
// wait for the resting screen to match the pre-probe baseline (Codex may
// redraw / scroll after space+DEL).
const defaultStabilizeAttempts = 20

// IO is the injectable snapshot/inject surface for Probe.
type IO struct {
	Snapshot func() (string, error)
	Inject   func(text string) error // no-submit; Probe sends " " then "\x7f"
	// Before is an optional resting snapshot. When non-empty, Probe skips the
	// first Snapshot call (avoids a duplicate SnapshotText under idle Tick).
	Before string
	// Settle is how long to wait after injecting the space before reading
	// after, and between post-DEL stabilize polls (0 → defaultProbeSettle).
	Settle time.Duration
}

// Probe types one space, waits Settle, snapshots, classifies with
// ExactlyOneMoreSpace, then always DELs the space and waits until Snapshot
// matches the pre-probe resting text again (or attempts are exhausted).
//
// exactly +1 draft space → Occupied
// any other / no change    → Empty
// snapshot fail            → Unknown
// inject fail              → Empty (cannot prove occupancy; callers that need
// a hard hold should check Ready/writable separately)
func Probe(io IO) Status {
	if io.Snapshot == nil || io.Inject == nil {
		return Unknown
	}
	before := io.Before
	if before == "" {
		var err error
		before, err = io.Snapshot()
		if err != nil {
			return Unknown
		}
	}
	if err := io.Inject(" "); err != nil {
		return Empty
	}

	settle := io.Settle
	if settle <= 0 {
		settle = defaultProbeSettle
	}
	time.Sleep(settle)

	after, err := io.Snapshot()
	status := Empty
	if err != nil {
		status = Unknown
	} else if after == before {
		status = Empty
	} else if ExactlyOneMoreSpace([]byte(before), []byte(after)) {
		status = Occupied
	}

	_ = io.Inject("\x7f")
	stabilizeAfterDEL(io, before, settle)
	return status
}

// stabilizeAfterDEL polls Snapshot until it equals the pre-probe resting
// baseline. Space+DEL can scroll/reflow the TUI; idle condition-1 needs the
// probe owner to leave the resting world as it found it.
func stabilizeAfterDEL(io IO, before string, settle time.Duration) {
	if io.Snapshot == nil || before == "" {
		return
	}
	if settle <= 0 {
		settle = defaultProbeSettle
	}
	for i := 0; i < defaultStabilizeAttempts; i++ {
		time.Sleep(settle)
		cur, err := io.Snapshot()
		if err != nil {
			return
		}
		if cur == before {
			return
		}
	}
}
