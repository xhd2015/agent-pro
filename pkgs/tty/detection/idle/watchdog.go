// Package idle decides exit-on-idle from snapshot stability + occupy probe.
package idle

import (
	"time"

	"github.com/xhd2015/agent-pro/pkgs/tty/detection/changed"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
)

// DefaultGrace is the post-SoftExit wait before Shutdown.
const DefaultGrace = 5 * time.Second

// Watchdog is the injectable keep-alive idle-exit state machine.
//
// Each Tick:
//  1. capture resting snapshot (pre-space baseline for this tick)
//  2. if changed vs last resting baseline → reset hits (start over)
//  3. if not Ready (when set) → reset hits (do not probe yet)
//  4. else probe occupy (space / compare / DEL), reusing resting snap as before
//  5. keep resting baseline as the pre-space snap (DEL restores; do not
//     Set(post-probe) — ephemeral placeholder collapse/restore must not look
//     like session activity on the next Tick)
//  6. if occupied → reset hits; Unknown after Ready is treated as empty
//  7. if QueueLen > 0 → reset hits
//  8. else count an idle hit; SoftExit after SamplesPerCycle consecutive hits
//     and continuous idle ≥ Timeout
type Watchdog struct {
	Timeout time.Duration
	Grace   time.Duration // 0 → DefaultGrace
	Now     func() time.Time

	// Snapshot returns the resting TTY snapshot text.
	Snapshot func() (string, error)
	// ProbeOccupied runs the space probe. When nil, Tick calls occupied.Probe
	// with the resting snap as Before + Snapshot + Inject.
	ProbeOccupied func() occupied.Status
	// Inject is used when ProbeOccupied is nil (no-submit " " / DEL).
	Inject func(text string) error
	// Ready is an optional hold (typically CheckWritable). nil → treat as ready.
	// Prevents SoftExit before the TUI can accept input / a draft.
	Ready func(snapshot string) bool
	// ReadyStatus when set is preferred over Ready and supplies ready_state for logs.
	ReadyStatus func(snapshot string) (ready bool, state string)
	// QueueLen is an optional hold. nil → treat as 0.
	QueueLen func() int

	SoftExit func()
	Shutdown func()

	// Log receives idle.jsonl events (armed/tick/reset/soft_exit/shutdown). nil → silent.
	Log func(Event)
	// SessionID is copied into log events when set.
	SessionID string

	armed     bool
	idleHits  int
	idleSince time.Time
	tracker   changed.Tracker
	exitAt    time.Time
	softDone  bool
	shutDone  bool
}

// Policy is the arming subset of idle-policy.json.
type Policy struct {
	ExitOnIdle  bool
	IdleTimeout time.Duration
}

// New copies cfg. Tick is a no-op when !found or !p.ExitOnIdle.
// Timeout comes from p.IdleTimeout when cfg.Timeout == 0.
func New(found bool, p Policy, cfg Watchdog) *Watchdog {
	w := cfg
	w.armed = found && p.ExitOnIdle
	if w.Timeout == 0 {
		w.Timeout = p.IdleTimeout
	}
	if w.Grace == 0 {
		w.Grace = DefaultGrace
	}
	if w.Now == nil {
		w.Now = time.Now
	}
	return &w
}

// IdleHits is the consecutive idle-check count (0..SamplesPerCycle).
func (w *Watchdog) IdleHits() int {
	if w == nil {
		return 0
	}
	return w.idleHits
}

// SoftDone reports whether SoftExit has already fired.
func (w *Watchdog) SoftDone() bool {
	return w != nil && w.softDone
}

// LogArmed writes idle.armed when the watchdog is armed and Log is set.
func (w *Watchdog) LogArmed() {
	if w == nil || !w.armed {
		return
	}
	w.emit(Event{
		Event:     EventArmed,
		SessionID: w.SessionID,
		Timeout:   w.Timeout.String(),
		Grace:     w.Grace.String(),
	})
}

// Tick advances one resting+occupy check.
func (w *Watchdog) Tick() {
	if w == nil || !w.armed {
		return
	}
	now := time.Time{}
	if w.Now != nil {
		now = w.Now()
	}

	snap, err := w.snapshotNow()
	if err != nil {
		w.resetHits(ResetSnapshotErr, "", now)
		return
	}
	if w.tracker.Note(snap) {
		w.resetHits(ResetChanged, snap, now)
		return
	}
	ready, readyState := w.readyOf(snap)
	if !ready {
		w.resetHits(ResetNotReady, snap, now)
		return
	}

	status := w.probe(snap)
	// Resting baseline stays the pre-space snap (Note already stored it).
	// Do not Set(post-probe): Codex idle hints collapse during the space probe
	// and redraw between ticks; locking baseline to the collapsed frame makes
	// the restored hint look like activity and forever resets idle hits.
	w.tracker.Set(snap)
	// After Ready, Unknown (e.g. mid-probe snapshot glitch) must not block exit.
	// Only a confirmed Occupied draft holds the session.
	if status == occupied.Occupied {
		w.resetHits(ResetOccupied, snap, now)
		return
	}
	q := w.queueLen()
	if q != 0 {
		w.resetHits(ResetQueue, snap, now)
		return
	}

	if w.idleSince.IsZero() {
		w.idleSince = now
	}
	if w.idleHits < SamplesPerCycle {
		w.idleHits++
	}
	hash, snapLen := SnapMeta(snap)
	readyTrue := true
	w.emit(Event{
		Event:      EventTick,
		SessionID:  w.SessionID,
		Hits:       w.idleHits,
		IdleSince:  formatTime(w.idleSince),
		Age:        now.Sub(w.idleSince).String(),
		Ready:      &readyTrue,
		ReadyState: readyState,
		Occupy:     string(status),
		Queue:      intPtr(q),
		SnapHash:   hash,
		SnapLen:    snapLen,
		Timeout:    w.Timeout.String(),
	})
	// SoftExit only after N consecutive idle checks AND continuous idle for
	// Timeout (matches "unchanged and not occupied for N").
	if !w.softDone && w.idleHits >= SamplesPerCycle && now.Sub(w.idleSince) >= w.Timeout {
		// Final occupy check: only a confirmed draft holds SoftExit.
		if st := w.probe(snap); st == occupied.Occupied {
			w.tracker.Set(snap)
			w.resetHits(ResetFinalOccupy, snap, now)
			return
		}
		w.tracker.Set(snap)
		w.softDone = true
		w.exitAt = now
		w.emit(Event{
			Event:      EventSoftExit,
			SessionID:  w.SessionID,
			Hits:       w.idleHits,
			IdleSince:  formatTime(w.idleSince),
			Age:        now.Sub(w.idleSince).String(),
			Ready:      &readyTrue,
			ReadyState: readyState,
			Occupy:     string(status),
			Queue:      intPtr(q),
			SnapHash:   hash,
			SnapLen:    snapLen,
			SnapTail:   SnapTail(snap, 0),
			Timeout:    w.Timeout.String(),
			Grace:      w.Grace.String(),
		})
		if w.SoftExit != nil {
			w.SoftExit()
		}
	}
	if w.softDone && !w.shutDone && now.Sub(w.exitAt) >= w.Grace {
		w.shutDone = true
		w.emit(Event{
			Event:     EventShutdown,
			SessionID: w.SessionID,
			Age:       now.Sub(w.exitAt).String(),
			Grace:     w.Grace.String(),
		})
		if w.Shutdown != nil {
			w.Shutdown()
		}
	}
}

// ForceShutdown fires Shutdown once (post-grace serve loop path).
func (w *Watchdog) ForceShutdown() {
	if w == nil || w.shutDone {
		return
	}
	w.shutDone = true
	now := time.Time{}
	if w.Now != nil {
		now = w.Now()
	}
	age := ""
	if !w.exitAt.IsZero() && !now.IsZero() {
		age = now.Sub(w.exitAt).String()
	}
	w.emit(Event{
		Event:     EventShutdown,
		SessionID: w.SessionID,
		Age:       age,
		Grace:     w.Grace.String(),
		Reason:    "force",
	})
	if w.Shutdown != nil {
		w.Shutdown()
	}
}

func (w *Watchdog) resetHits(reason, snap string, now time.Time) {
	before := w.idleHits
	idleSince := w.idleSince
	w.idleHits = 0
	w.idleSince = time.Time{}

	ev := Event{
		Event:      EventReset,
		SessionID:  w.SessionID,
		Reason:     reason,
		HitsBefore: before,
		IdleSince:  formatTime(idleSince),
		Timeout:    w.Timeout.String(),
	}
	if !idleSince.IsZero() && !now.IsZero() {
		ev.Age = now.Sub(idleSince).String()
	}
	if snap != "" {
		hash, snapLen := SnapMeta(snap)
		ev.SnapHash = hash
		ev.SnapLen = snapLen
		if reason == ResetChanged {
			ev.SnapTail = SnapTail(snap, 0)
		}
		if reason == ResetNotReady {
			_, state := w.readyOf(snap)
			ev.ReadyState = state
			readyFalse := false
			ev.Ready = &readyFalse
		}
	}
	w.emit(ev)
}

func (w *Watchdog) readyOf(snap string) (ready bool, state string) {
	if w.ReadyStatus != nil {
		return w.ReadyStatus(snap)
	}
	if w.Ready != nil {
		return w.Ready(snap), ""
	}
	return true, ""
}

func (w *Watchdog) emit(e Event) {
	if w == nil || w.Log == nil {
		return
	}
	if e.TS.IsZero() && w.Now != nil {
		e.TS = w.Now()
	}
	w.Log(e)
}

func (w *Watchdog) snapshotNow() (string, error) {
	if w.Snapshot == nil {
		return "", errNoSnapshot
	}
	return w.Snapshot()
}

func (w *Watchdog) probe(resting string) occupied.Status {
	if w.ProbeOccupied != nil {
		return w.ProbeOccupied()
	}
	if w.Snapshot == nil || w.Inject == nil {
		return occupied.Unknown
	}
	return occupied.Probe(occupied.IO{
		Before:   resting,
		Snapshot: w.Snapshot,
		Inject:   w.Inject,
	})
}

func (w *Watchdog) queueLen() int {
	if w.QueueLen == nil {
		return 0
	}
	return w.QueueLen()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func intPtr(n int) *int { return &n }

type snapshotError string

func (e snapshotError) Error() string { return string(e) }

const errNoSnapshot = snapshotError("idle: Snapshot not configured")
