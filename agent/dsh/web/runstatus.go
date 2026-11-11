package web

import (
	"context"
	"sync"
	"time"
)

// DefaultRunStatusTTL bounds how stale a cached run-status answer may be. The
// daemon answers every loaded Session in one round trip, so a short window
// collapses a burst of per-Session lookups into one RPC while still following a
// turn that starts or ends within the second.
const DefaultRunStatusTTL = time.Second

// runStatusTimeout caps one status RPC. The daemon answers from memory, but a
// saturated event loop (a page reload, a large session list, a cold registry)
// can queue one reply for a long while, and an aborted batch turns every
// Session into an error the caller only gets to see as a flash. A minute is
// long enough to outlast a stall and short enough to bound a dead daemon; a
// caller that cannot wait sets RunStatus.Timeout.
const runStatusTimeout = 60 * time.Second

// Status asks the daemon which Sessions are executing a turn.
//
// A nil ids asks about every Session the daemon has loaded; an empty slice asks
// about none. An id is absent from the result when the daemon never loaded it,
// denies it to this caller, or does not know it, so absence means unknown here
// rather than idle — a Session opened by a plain `dsh` TUI is not in the
// daemon's registry at all.
func (c Client) Status(ctx context.Context, ids []string) (map[string]bool, error) {
	request := map[string]any{}
	if ids != nil {
		request["ids"] = ids
	}
	var value struct {
		Running map[string]bool `json:"running"`
	}
	if err := c.rpc(ctx, "session/status", map[string]any{"request": request}, &value); err != nil {
		return nil, err
	}
	if value.Running == nil {
		return map[string]bool{}, nil
	}
	return value.Running, nil
}

// RunStatus is the liveness hook a phase reader installs: Running reports
// whether one Session is executing a turn, and whether the daemon could answer
// at all.
//
// It is shaped to match sessionactivity.RunStatus without importing it, so the
// activity package keeps its dependency-free contract and this package stays
// transport-only. Install it with sessionactivity.SetDefaultRunStatus.
//
// One RPC per TTL covers every Session, because the daemon's answer is batch.
// A failed refresh reports unknown for everything instead of replaying the last
// good map: a dead daemon must degrade to the caller's process-lease liveness,
// never to a stale "running".
type RunStatus struct {
	Client Client
	TTL    time.Duration
	// Timeout overrides how long one status RPC may take. Zero uses
	// runStatusTimeout.
	Timeout time.Duration
	// Now overrides the clock in tests. Nil uses time.Now.
	Now func() time.Time

	mu        sync.Mutex
	fetchedAt time.Time
	known     bool
	running   map[string]bool
}

// NewRunStatus returns the hook for one DSH web instance, refreshing at most
// once per DefaultRunStatusTTL.
func NewRunStatus(client Client) *RunStatus {
	return &RunStatus{Client: client}
}

// Answer is the session/status answer about one id: whether the RPC succeeded,
// whether the id was in the map, whether that entry is running, and when the
// batch was sampled.
//
// At is the instant the request was sent, so it never claims to be fresher than
// the answer. A cache hit reports the batch's instant, not the lookup's: that is
// what lets a reader decide whether the answer could have seen a turn the log
// already has.
type Answer struct {
	Running  bool
	Present  bool
	Answered bool
	At       time.Time
}

// Lookup is the session/status answer for one id.
func (r *RunStatus) Lookup(sessionID string) Answer {
	current, at, answered := r.current()
	answer := Answer{At: at}
	if !answered {
		return answer
	}
	value, present := current[sessionID]
	answer.Running = value
	answer.Present = present
	answer.Answered = true
	return answer
}

// Running reports whether the daemon's registry has sessionID executing a turn.
// known is true only when the RPC succeeded and named this id.
func (r *RunStatus) Running(sessionID string) (running bool, known bool) {
	answer := r.Lookup(sessionID)
	return answer.Running, answer.Answered && answer.Present
}

// current returns the cached batch answer, refreshing it once per TTL.
//
// The lock is held across the RPC on purpose: concurrent readers then wait for
// one batched answer instead of each sending its own request, and the wait is
// capped by runStatusTimeout.
func (r *RunStatus) current() (map[string]bool, time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if r.Now != nil {
		now = r.Now()
	}
	ttl := r.TTL
	if ttl <= 0 {
		ttl = DefaultRunStatusTTL
	}
	// A failed refresh is cached for the same window: a daemon that is down must
	// not be retried once per Session lookup.
	if !r.fetchedAt.IsZero() && now.Sub(r.fetchedAt) < ttl {
		return r.running, r.fetchedAt, r.known
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = runStatusTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	running, err := r.Client.Status(ctx, nil)
	// The request instant, not the reply instant: the daemon decides while the
	// answer is in flight, so this is the earliest moment the answer is guaranteed
	// to reflect. Sampling later would make the answer look fresher than it is.
	r.fetchedAt = now
	if err != nil {
		r.known = false
		r.running = nil
		return nil, r.fetchedAt, false
	}
	r.known = true
	r.running = running
	return running, r.fetchedAt, true
}
