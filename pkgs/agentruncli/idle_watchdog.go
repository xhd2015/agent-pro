package agentruncli

import "time"

const defaultIdleGrace = 5 * time.Second

// idleWatchFirstDelay is the earliest first snapshot when IdleTimeout >= this.
// Timeouts shorter than this (e.g. 10s probes) sample immediately.
const idleWatchFirstDelay = 30 * time.Second

// idleWatchSamplesPerCycle is the max snapshots in one idle-exit cycle.
const idleWatchSamplesPerCycle = 3

// IdleSample is one watchdog observation of TTY + queue idleness.
type IdleSample struct {
	Sendable bool
	Screen   string // idle|busy|starting|modal|unknown
	InputBox string // empty|occupied|unknown
	QueueLen int
	// LogFound / LogBytes are the Codex rollout jsonl Stat.Size gate.
	// Missing file → LogFound=false and Tick skips this gate.
	LogFound bool
	LogBytes int64
}

// SampleIsIdle is true only when sendable, screen idle, box empty, and queue empty.
// Log size is a separate Tick gate (not part of this predicate).
func SampleIsIdle(s IdleSample) bool {
	return s.Sendable && s.Screen == "idle" && s.InputBox == "empty" && s.QueueLen == 0
}

// IdleWatchdog is the injectable keep-alive idle-exit state machine.
type IdleWatchdog struct {
	Timeout  time.Duration
	Grace    time.Duration // 0 → 5s
	Now      func() time.Time
	Sample   func() IdleSample
	SoftExit func()
	Shutdown func()

	armed     bool
	idleHits  int
	idleSince time.Time
	exitAt    time.Time
	softDone  bool
	shutDone  bool

	haveLogSize bool
	lastLogSize int64
}

// NewIdleWatchdog copies cfg. Tick is a no-op when !found or !p.ExitOnIdle.
// Timeout comes from p.IdleTimeout when cfg.Timeout == 0.
func NewIdleWatchdog(found bool, p IdlePolicy, cfg IdleWatchdog) *IdleWatchdog {
	w := cfg
	w.armed = found && p.ExitOnIdle
	if w.Timeout == 0 {
		w.Timeout = p.IdleTimeout
	}
	if w.Grace == 0 {
		w.Grace = defaultIdleGrace
	}
	if w.Now == nil {
		w.Now = time.Now
	}
	return &w
}

// IdleWatchSchedule is the serve-loop sleep plan for one cycle: first delay,
// then two gaps. Timeouts < 30s start immediately (0, T/2, T). Longer
// timeouts wait 30s first (30s, 30s+(T-30s)/2, T). At most 3 snapshots.
func IdleWatchSchedule(timeout time.Duration) (first, gap time.Duration) {
	if timeout <= 0 {
		return 0, 0
	}
	first = idleWatchFirstDelay
	if timeout < idleWatchFirstDelay {
		first = 0
	}
	remain := timeout - first
	if remain < 0 {
		remain = 0
	}
	return first, remain / 2
}

// Tick advances one sample. SoftExit fires once after idleWatchSamplesPerCycle
// consecutive idle samples; Shutdown fires once after grace on a later Tick.
// Non-idle (chrome or jsonl size change) clears the hit count.
// The serve loop sleeps between Ticks.
func (w *IdleWatchdog) Tick() {
	if w == nil || !w.armed {
		return
	}
	now := time.Time{}
	if w.Now != nil {
		now = w.Now()
	}
	var sample IdleSample
	if w.Sample != nil {
		sample = w.Sample()
	}
	logGrew := w.noteLogSize(sample)
	if !SampleIsIdle(sample) || logGrew {
		w.idleHits = 0
		w.idleSince = time.Time{}
		return
	}
	if w.idleSince.IsZero() {
		w.idleSince = now
	}
	if w.idleHits < idleWatchSamplesPerCycle {
		w.idleHits++
	}
	if !w.softDone && w.idleHits >= idleWatchSamplesPerCycle {
		w.softDone = true
		w.exitAt = now
		if w.SoftExit != nil {
			w.SoftExit()
		}
	}
	if w.softDone && !w.shutDone && now.Sub(w.exitAt) >= w.Grace {
		w.shutDone = true
		if w.Shutdown != nil {
			w.Shutdown()
		}
	}
}

// noteLogSize records rollout size. First observation is a baseline (not a
// change). Later Stat.Size != last → grew. Missing file skips the gate.
func (w *IdleWatchdog) noteLogSize(sample IdleSample) (grew bool) {
	if w == nil || !sample.LogFound {
		return false
	}
	if !w.haveLogSize {
		w.lastLogSize = sample.LogBytes
		w.haveLogSize = true
		return false
	}
	if sample.LogBytes != w.lastLogSize {
		w.lastLogSize = sample.LogBytes
		return true
	}
	return false
}

func (w *IdleWatchdog) forceShutdown() {
	if w == nil || w.shutDone {
		return
	}
	w.shutDone = true
	if w.Shutdown != nil {
		w.Shutdown()
	}
}
