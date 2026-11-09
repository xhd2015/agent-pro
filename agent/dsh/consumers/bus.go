package consumers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// recordSchema versions the on-disk cache layout. A bump starts every session
// from the stream origin without touching the session logs.
const recordSchema = 1

// CacheSuffix is the directory under the user cache directory that holds the
// per-session records. Deleting it only costs a re-read.
const CacheSuffix = "agent-pro/session-consumers/v1"

// Default read budget per session.
const (
	DefaultWindow     int64 = 1 << 20
	DefaultMaxWindow  int64 = 32 << 20
	DefaultMaxEvents        = 1024
	DefaultMaxBatches       = 64
	DefaultBudget           = 250 * time.Millisecond
)

// Options configure a Bus.
type Options struct {
	// Dir overrides where records are cached. Empty uses the user cache dir.
	Dir string
	// Consumers is the fold set. Version a consumer to replay its own state.
	Consumers []Consumer
	// Window is the compressed bytes read per attempt (default 1 MiB).
	Window int64
	// MaxWindow caps growth for a single large frame (default 32 MiB).
	MaxWindow int64
	// MaxEvents stops a batch near this many rows (default 1024).
	MaxEvents int
	// MaxBatches bounds the batches folded by one Sync call (default 64).
	MaxBatches int
	// Budget bounds the wall time one Sync call spends reading (default 250ms).
	Budget time.Duration
	// Parallelism is the worker count for SyncAll (default min(8, NumCPU)).
	Parallelism int
	// MaxOpenStreams bounds concurrent open streams (default 4).
	MaxOpenStreams int
}

// Snapshot is the state after one Sync.
type Snapshot struct {
	SessionID     string
	ArtifactPath  string
	FormatVersion int
	// Size is the artifact size observed by this call.
	Size int64
	// LastSeq is the last row folded by this session's stream.
	LastSeq int64
	// CaughtUp reports that no durable row remained after this call. Callers
	// should only trust an absent fact (no pending question, no running turn)
	// when this is true.
	CaughtUp bool
	// ReadBytes is how many bytes this call advanced through.
	ReadBytes int64
	// Folded is how many events were handed to a consumer this call.
	Folded int64
	// Skipped is how many events a consumer already had and skipped. It is
	// non-zero when a version bump replayed the stream for one consumer.
	Skipped int64
	// Values holds each consumer's raw state, keyed by consumer ID.
	Values map[string][]byte
}

// Value decodes one consumer's state.
func (s Snapshot) Value(consumer string, target any) bool {
	raw, ok := s.Values[consumer]
	if !ok || len(raw) == 0 {
		return false
	}
	return json.Unmarshal(raw, target) == nil
}

// Stats is a snapshot of bus counters.
type Stats struct {
	Syncs     int64 `json:"syncs"`
	Streams   int64 `json:"streams"`
	CacheHits int64 `json:"cacheHits"`
	BytesRead int64 `json:"bytesRead"`
	Events    int64 `json:"events"`
	Folded    int64 `json:"folded"`
	Skipped   int64 `json:"skipped"`
	Resets    int64 `json:"resets"`
	Writes    int64 `json:"writes"`
	Errors    int64 `json:"errors"`
}

type counters struct {
	syncs     atomic.Int64
	streams   atomic.Int64
	cacheHits atomic.Int64
	bytesRead atomic.Int64
	events    atomic.Int64
	folded    atomic.Int64
	skipped   atomic.Int64
	resets    atomic.Int64
	writes    atomic.Int64
	errors    atomic.Int64
}

// Bus keeps one stream per session and folds it into every registered
// consumer. It is safe for concurrent use; concurrent Sync calls for the same
// session serialize behind that session's stream.
type Bus struct {
	sessions  *sessions.Store
	cache     recordStore
	consumers []Consumer
	// types maps consumer ID to the event types it watches.
	types map[string]map[string]bool
	// interest is the union of every watched type.
	interest map[string]bool

	window     int64
	maxWindow  int64
	maxEvents  int
	maxBatches int
	budget     time.Duration
	parallel   int

	mu      sync.Mutex
	records map[string]*record

	locks sync.Map // session ID -> *sync.Mutex
	sem   chan struct{}

	counters counters
}

// New builds a bus over a session store.
func New(store *sessions.Store, opts Options) (*Bus, error) {
	if store == nil {
		return nil, fmt.Errorf("consumers: session store is required")
	}
	if err := validate(opts.Consumers); err != nil {
		return nil, fmt.Errorf("consumers: %w", err)
	}
	dir := opts.Dir
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("consumers: locate cache dir: %w", err)
		}
		dir = filepath.Join(base, filepath.FromSlash(CacheSuffix))
	}
	bus := &Bus{
		sessions:   store,
		cache:      recordStore{dir: dir},
		consumers:  opts.Consumers,
		types:      make(map[string]map[string]bool, len(opts.Consumers)),
		interest:   make(map[string]bool),
		window:     opts.Window,
		maxWindow:  opts.MaxWindow,
		maxEvents:  opts.MaxEvents,
		maxBatches: opts.MaxBatches,
		budget:     opts.Budget,
		parallel:   opts.Parallelism,
		records:    make(map[string]*record),
	}
	if bus.window <= 0 {
		bus.window = DefaultWindow
	}
	if bus.maxWindow < bus.window {
		bus.maxWindow = DefaultMaxWindow
	}
	if bus.maxEvents <= 0 {
		bus.maxEvents = DefaultMaxEvents
	}
	if bus.maxBatches <= 0 {
		bus.maxBatches = DefaultMaxBatches
	}
	if bus.budget <= 0 {
		bus.budget = DefaultBudget
	}
	if bus.parallel <= 0 {
		bus.parallel = min(8, runtime.NumCPU())
	}
	for _, consumer := range opts.Consumers {
		watched := make(map[string]bool)
		for _, typ := range consumer.Types() {
			watched[typ] = true
			bus.interest[typ] = true
		}
		bus.types[consumer.ID()] = watched
	}
	open := opts.MaxOpenStreams
	if open <= 0 {
		open = 4
	}
	bus.sem = make(chan struct{}, open)
	return bus, nil
}

// Consumers lists the registered consumer IDs.
func (b *Bus) Consumers() []string {
	return sortedIDs(b.consumers)
}

// CacheDir is where records are stored.
func (b *Bus) CacheDir() string { return b.cache.dir }

// Stats reports bus counters.
func (b *Bus) Stats() Stats {
	return Stats{
		Syncs:     b.counters.syncs.Load(),
		Streams:   b.counters.streams.Load(),
		CacheHits: b.counters.cacheHits.Load(),
		BytesRead: b.counters.bytesRead.Load(),
		Events:    b.counters.events.Load(),
		Folded:    b.counters.folded.Load(),
		Skipped:   b.counters.skipped.Load(),
		Resets:    b.counters.resets.Load(),
		Writes:    b.counters.writes.Load(),
		Errors:    b.counters.errors.Load(),
	}
}

// Peek returns what the folds already know about one session.
//
// Peek never reads the log and never waits on a stream: it answers from the
// in-memory record, or from the one on disk when this process has not touched
// the session yet. A request handler fills a header or a footer with it instead
// of paying for a cold fold, where reading a large log costs seconds.
//
// The second result is false when nothing usable is cached, and also when the
// record predates a registered consumer: its absent state would read as zero.
// CaughtUp reports whether the rows behind the record were folded to the end of
// the artifact; a caught-up record can still be behind the artifact's current
// size, so a caller that needs fresh numbers compares Size with the log.
func (b *Bus) Peek(id string) (Snapshot, bool) {
	if id == "" {
		return Snapshot{}, false
	}
	lock := b.sessionLock(id)
	lock.Lock()
	defer lock.Unlock()
	rec, err := b.load(id)
	if err != nil || !rec.covers(b.consumers) {
		return Snapshot{}, false
	}
	b.remember(id, rec)
	return rec.snapshot(0, 0, 0), true
}

// Sync brings one session up to date and returns its consumers' state.
//
// The stream resumes from the cached frame boundary and reads only new bytes.
// When a consumer's Version changed, that consumer's state restarts from the
// origin and the stream is re-read while the unchanged consumers skip rows they
// already folded.
func (b *Bus) Sync(id string) (Snapshot, error) {
	if id == "" {
		return Snapshot{}, fmt.Errorf("consumers: session id is required")
	}
	lock := b.sessionLock(id)
	lock.Lock()
	defer lock.Unlock()

	b.sem <- struct{}{}
	defer func() { <-b.sem }()

	b.counters.syncs.Add(1)
	rec, err := b.load(id)
	if err != nil {
		b.counters.errors.Add(1)
		return Snapshot{}, err
	}
	ref, err := b.sessions.StatLog(id)
	if err != nil {
		b.counters.errors.Add(1)
		if rec != nil {
			return rec.snapshot(0, 0, 0), err
		}
		return Snapshot{SessionID: id}, err
	}
	if !rec.valid(id, ref) {
		if rec != nil {
			b.counters.resets.Add(1)
		}
		rec = newRecord(id, ref, b.consumers)
	}
	if rec.ensure(b.consumers) {
		b.counters.resets.Add(1)
	}

	// The artifact did not change: no file read at all.
	if rec.Next == ref.Size && rec.CaughtUp {
		b.counters.cacheHits.Add(1)
		b.remember(id, rec)
		return rec.snapshot(0, 0, 0), nil
	}

	var (
		readBytes int64
		folded    int64
		skipped   int64
		dirty     bool
		resynced  bool
	)
	deadline := time.Now().Add(b.budget)
	for batchIndex := 0; batchIndex < b.maxBatches; batchIndex++ {
		if readBytes > 0 && time.Now().After(deadline) {
			break
		}
		from := rec.Next
		batch, err := b.sessions.ReadLog(id, sessions.ReadLogOptions{
			From:      from,
			Window:    b.window,
			MaxWindow: b.maxWindow,
			MaxEvents: b.maxEvents,
			Interest:  func(typ string) bool { return b.interest[typ] },
		})
		if err != nil {
			b.counters.errors.Add(1)
			b.forget(id)
			return rec.snapshot(readBytes, folded, skipped), err
		}
		if batch.Rewound || !batch.Ref.SameArtifact(sessions.LogRef{
			Path:          rec.ArtifactPath,
			Encoding:      batch.Ref.Encoding,
			FormatVersion: rec.FormatVersion,
		}) {
			// A different artifact now occupies the path: its rows are a
			// different stream, so every consumer restarts.
			rec = newRecord(id, batch.Ref, b.consumers)
			b.counters.resets.Add(1)
			continue
		}
		if len(batch.Events) == 0 {
			rec.Next = batch.Next
			rec.Size = batch.Ref.Size
			rec.ModTimeUnixNano = batch.Ref.ModTimeUnixNano
			rec.CaughtUp = batch.Done
			dirty = true
			break
		}
		// The origin read accepts any first seq: a log that was rotated or
		// truncated can start mid-sequence. Every later read requires a dense
		// sequence, which is what catches an artifact rewritten under the
		// cache.
		expected := rec.LastSeq + 1
		if rec.LastSeq < 0 {
			expected = batch.Events[0].Seq
		}
		if got, want, bad := seqGap(expected, batch.Events); bad {
			if resynced {
				b.counters.errors.Add(1)
				b.forget(id)
				return rec.snapshot(readBytes, folded, skipped),
					fmt.Errorf("consumers: session %s: event %d where %d was expected", id, got, want)
			}
			// The artifact was rewritten under the cache. Start over once.
			resynced = true
			rec = newRecord(id, batch.Ref, b.consumers)
			b.counters.resets.Add(1)
			continue
		}
		batchFolded, batchSkipped, err := b.apply(rec, batch.Events, batch.Ref.Size)
		if err != nil {
			b.counters.errors.Add(1)
			b.forget(id)
			return rec.snapshot(readBytes, folded, skipped), fmt.Errorf("consumers: session %s: %w", id, err)
		}
		folded += batchFolded
		skipped += batchSkipped
		readBytes += batch.Next - from
		rec.Next = batch.Next
		rec.LastSeq = batch.Events[len(batch.Events)-1].Seq
		rec.Size = batch.Ref.Size
		rec.ModTimeUnixNano = batch.Ref.ModTimeUnixNano
		rec.CaughtUp = batch.Done
		dirty = true
		if batch.Done || !batch.More {
			break
		}
	}
	if dirty {
		b.counters.writes.Add(1)
		// A cache that cannot be written is not fatal: the states stay in
		// memory and the next process pays for a re-read.
		_ = b.cache.save(rec)
	}
	b.remember(id, rec)
	b.counters.bytesRead.Add(readBytes)
	b.counters.events.Add(folded + skipped)
	b.counters.folded.Add(folded)
	b.counters.skipped.Add(skipped)
	if readBytes > 0 {
		b.counters.streams.Add(1)
	}
	return rec.snapshot(readBytes, folded, skipped), nil
}

// SyncAll syncs many sessions and returns the snapshots it completed. Failures
// are counted, not returned: one unreadable session must not fail a listing.
func (b *Bus) SyncAll(ids []string) map[string]Snapshot {
	snapshots := make(map[string]Snapshot, len(ids))
	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		work = make(chan string)
	)
	workers := b.parallel
	if workers > len(ids) {
		workers = len(ids)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range work {
				snapshot, err := b.Sync(id)
				if err != nil {
					continue
				}
				mu.Lock()
				snapshots[id] = snapshot
				mu.Unlock()
			}
		}()
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		work <- id
	}
	close(work)
	wg.Wait()
	return snapshots
}

// Forget drops one session's cached state, in memory and on disk.
func (b *Bus) Forget(id string) error {
	lock := b.sessionLock(id)
	lock.Lock()
	defer lock.Unlock()
	b.forget(id)
	return b.cache.remove(id)
}

// seqGap reports the first row of a batch that does not continue the sequence
// that starts at expected.
func seqGap(expected int64, events []sessions.LogEvent) (got, want int64, bad bool) {
	for i := range events {
		if want := expected + int64(i); events[i].Seq != want {
			return events[i].Seq, want, true
		}
	}
	return 0, 0, false
}

// apply folds a batch into every consumer. The states are decoded once per
// batch and re-encoded once, so a batch costs one parse per changed consumer
// rather than one per event.
func (b *Bus) apply(rec *record, events []sessions.LogEvent, size int64) (folded, skipped int64, err error) {
	live := make(map[string]any, len(b.consumers))
	for i := range events {
		event := Event{
			Seq:  events[i].Seq,
			Type: events[i].Type,
			Time: events[i].Time,
			Data: events[i].Data,
		}
		at := Point{Seq: event.Seq, Size: size}
		for _, consumer := range b.consumers {
			id := consumer.ID()
			entry := rec.Consumers[id]
			// Already folded by this consumer: the stream was re-read for a
			// version bump, so skip until it catches up.
			if event.Seq <= entry.Seq {
				skipped++
				continue
			}
			entry.Seq = event.Seq
			rec.Consumers[id] = entry
			if !b.types[id][event.Type] {
				continue
			}
			state, ok := live[id]
			if !ok {
				state = consumer.New()
				if len(entry.State) > 0 {
					if decodeErr := json.Unmarshal(entry.State, state); decodeErr != nil {
						state = consumer.New()
					}
				}
				live[id] = state
			}
			if foldErr := consumer.Fold(state, event, at); foldErr != nil {
				return folded, skipped, foldErr
			}
			folded++
		}
	}
	for id, state := range live {
		raw, marshalErr := json.Marshal(state)
		if marshalErr != nil {
			return folded, skipped, marshalErr
		}
		entry := rec.Consumers[id]
		entry.State = raw
		rec.Consumers[id] = entry
	}
	return folded, skipped, nil
}

func (b *Bus) sessionLock(id string) *sync.Mutex {
	if lock, ok := b.locks.Load(id); ok {
		return lock.(*sync.Mutex)
	}
	lock, _ := b.locks.LoadOrStore(id, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (b *Bus) load(id string) (*record, error) {
	b.mu.Lock()
	rec := b.records[id]
	b.mu.Unlock()
	if rec != nil {
		return rec, nil
	}
	return b.cache.load(id)
}

func (b *Bus) remember(id string, rec *record) {
	b.mu.Lock()
	b.records[id] = rec
	b.mu.Unlock()
}

func (b *Bus) forget(id string) {
	b.mu.Lock()
	delete(b.records, id)
	b.mu.Unlock()
}
