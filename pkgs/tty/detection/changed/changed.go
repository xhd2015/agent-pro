// Package changed detects session activity by comparing resting snapshots.
package changed

import "bytes"

// Equal reports whether a and b are identical byte-for-byte (newlines count).
func Equal(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// Changed reports whether after differs from before.
func Changed(before, after []byte) bool {
	return !Equal(before, after)
}
// Tracker remembers the last resting snapshot. The first Note is a baseline
// only (not a change).
type Tracker struct {
	have bool
	last string
}

// Note records now. Returns whether the session changed since the last Note.
// First call always returns false and stores the baseline.
func (t *Tracker) Note(now string) (changed bool) {
	if t == nil {
		return false
	}
	if !t.have {
		t.last = now
		t.have = true
		return false
	}
	if Equal([]byte(t.last), []byte(now)) {
		return false
	}
	t.last = now
	return true
}

// Reset clears the baseline so the next Note baselines again.
func (t *Tracker) Reset() {
	if t == nil {
		return
	}
	t.have = false
	t.last = ""
}

// Set stores now as the resting baseline without reporting a change.
// Used after an occupy probe so inject/DEL side effects are not treated as
// session activity on the next sample.
func (t *Tracker) Set(now string) {
	if t == nil {
		return
	}
	t.last = now
	t.have = true
}
