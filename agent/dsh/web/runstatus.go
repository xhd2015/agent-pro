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

// runStatusTimeout caps one status RPC. The daemon answers from memory, so a
// slow reply means it is saturated or gone, and the caller's lease fallback is
// better than a long stall.
const runStatusTimeout = 3 * time.Second

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

// Running reports whether the daemon's registry has sessionID executing a turn.
// known is false when the daemon did not name the Session or could not answer.
func (r *RunStatus) Running(sessionID string) (running bool, known bool) {
	current, answered := r.current()
	if !answered {
		return false, false
	}
	value, present := current[sessionID]
	return value, present
}

// current returns the cached batch answer, refreshing it once per TTL.
//
// The lock is held across the RPC on purpose: concurrent readers then wait for
// one batched answer instead of each sending its own request, and the wait is
// capped by runStatusTimeout.
func (r *RunStatus) current() (map[string]bool, bool) {
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
		return r.running, r.known
	}
	ctx, cancel := context.WithTimeout(context.Background(), runStatusTimeout)
	defer cancel()
	running, err := r.Client.Status(ctx, nil)
	r.fetchedAt = now
	if err != nil {
		r.known = false
		r.running = nil
		return nil, false
	}
	r.known = true
	r.running = running
	return running, true
}
