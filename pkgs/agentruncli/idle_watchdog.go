package agentruncli

import "time"

const defaultIdleGrace = 5 * time.Second

// IdleSample is one watchdog observation of TTY + queue idleness.
type IdleSample struct {
	Sendable bool
	Screen   string // idle|busy|starting|modal|unknown
	InputBox string // empty|occupied|unknown
	QueueLen int
}

// SampleIsIdle is true only when sendable, screen idle, box empty, and queue empty.
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
	idleSince time.Time
	exitAt    time.Time
	softDone  bool
	shutDone  bool
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

// Tick advances the idle clock one sample. SoftExit fires once at timeout;
// Shutdown fires once after grace. Non-idle clears idleSince only.
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
	if !SampleIsIdle(sample) {
		w.idleSince = time.Time{}
		return
	}
	if w.idleSince.IsZero() {
		w.idleSince = now
	}
	if !w.softDone && now.Sub(w.idleSince) >= w.Timeout {
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
