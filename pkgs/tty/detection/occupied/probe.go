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

// IO is the injectable snapshot/inject surface for Probe.
type IO struct {
	Snapshot func() (string, error)
	Inject   func(text string) error // no-submit; Probe sends " " then "\x7f"
	// Before is an optional resting snapshot. When non-empty, Probe skips the
	// first Snapshot call (avoids a duplicate SnapshotText under idle Tick).
	Before string
	// Settle is how long to wait after injecting the space before reading
	// after (0 → defaultProbeSettle).
	Settle time.Duration
}

// Probe types one space, waits Settle, snapshots, classifies with
// ExactlyOneMoreSpace, then always DELs the space.
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
	defer func() {
		_ = io.Inject("\x7f")
	}()

	settle := io.Settle
	if settle <= 0 {
		settle = defaultProbeSettle
	}
	time.Sleep(settle)

	after, err := io.Snapshot()
	if err != nil {
		return Unknown
	}
	if after == before {
		return Empty
	}
	if ExactlyOneMoreSpace([]byte(before), []byte(after)) {
		return Occupied
	}
	return Empty
}
