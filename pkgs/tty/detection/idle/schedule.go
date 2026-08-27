package idle

import "time"

// FirstDelay is the earliest first snapshot when Timeout >= this.
// Timeouts shorter than this (e.g. 10s probes) sample immediately.
const FirstDelay = 30 * time.Second

// SamplesPerCycle is the max resting+occupy checks in one idle-exit cycle.
const SamplesPerCycle = 3

// Schedule is the serve-loop sleep plan for one cycle: first delay, then two
// gaps. Timeouts < 30s start immediately (0, T/2, T). Longer timeouts wait 30s
// first (30s, 30s+(T-30s)/2, T). At most 3 snapshots.
func Schedule(timeout time.Duration) (first, gap time.Duration) {
	if timeout <= 0 {
		return 0, 0
	}
	first = FirstDelay
	if timeout < FirstDelay {
		first = 0
	}
	remain := timeout - first
	if remain < 0 {
		remain = 0
	}
	return first, remain / 2
}
