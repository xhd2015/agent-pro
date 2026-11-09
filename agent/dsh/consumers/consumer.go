// Package consumers follows DSH session logs incrementally.
//
// A session log is append-only, so a consumer only needs the rows it has not
// folded yet. Bus keeps one stream per session, reads it in bounded windows,
// and hands each batch to every Consumer; each consumer's small state plus its
// position is cached per session, so the next call reads only new bytes.
//
// Consumers are versioned. Bumping a consumer's Version invalidates only that
// consumer's cached state: the session stream is re-read from the start and the
// consumers whose version did not change skip every row they already folded.
package consumers

import (
	"fmt"
	"sort"
)

// Event is one durable row of a session log.
type Event struct {
	Seq  int64
	Type string
	Time int64
	// Data is the raw payload, or nil when the type is outside the consumer's
	// Types set.
	Data []byte
}

// Point is the stream position a fold runs at.
type Point struct {
	// Seq is the row's sequence number.
	Seq int64
	// Size is the artifact size as of this read.
	Size int64
}

// Consumer folds a log into a small state that is persisted between runs.
//
// Fold must be a pure function of the events it has seen plus the state: it may
// keep counters, the latest value of a field, or a bounded set, but never the
// transcript. A state that is cheap to serialize is what makes a cache hit
// cheap.
type Consumer interface {
	// ID names the consumer. It is the cache key, so it must be stable.
	ID() string
	// Version of the fold. Bump it whenever Fold changes meaning.
	Version() int
	// Types lists the event types Fold reads. Everything else is skipped without
	// decoding its payload.
	Types() []string
	// New returns a pointer to a zero state.
	New() any
	// Fold applies one event to the state. state is the value New returned.
	Fold(state any, ev Event, at Point) error
}

// Typed adapts a struct state to Consumer.
type Typed[S any] struct {
	// Name is the consumer ID.
	Name string
	// Ver is the fold version.
	Ver int
	// Watch is the set of event types Apply reads.
	Watch []string
	// Init may seed defaults; it runs before the first fold.
	Init func(*S)
	// Apply folds one event.
	Apply func(*S, Event, Point) error
}

var _ Consumer = (*Typed[struct{}])(nil)

// ID implements Consumer.
func (t *Typed[S]) ID() string { return t.Name }

// Version implements Consumer.
func (t *Typed[S]) Version() int { return t.Ver }

// Types implements Consumer.
func (t *Typed[S]) Types() []string { return t.Watch }

// New implements Consumer.
func (t *Typed[S]) New() any {
	state := new(S)
	if t.Init != nil {
		t.Init(state)
	}
	return state
}

// Fold implements Consumer.
func (t *Typed[S]) Fold(state any, ev Event, at Point) error {
	typed, ok := state.(*S)
	if !ok {
		return fmt.Errorf("consumer %s: unexpected state %T", t.Name, state)
	}
	if t.Apply == nil {
		return fmt.Errorf("consumer %s: no fold", t.Name)
	}
	return t.Apply(typed, ev, at)
}

// Defaults returns the built-in consumers.
func Defaults() []Consumer {
	return []Consumer{
		Title(),
		Outcome(),
		Question(),
		Usage(),
	}
}

// validate rejects a consumer set that could not produce a stable cache.
func validate(consumers []Consumer) error {
	seen := make(map[string]bool, len(consumers))
	for _, consumer := range consumers {
		if consumer == nil {
			return fmt.Errorf("consumer is nil")
		}
		id := consumer.ID()
		if id == "" {
			return fmt.Errorf("consumer without an id")
		}
		if consumer.Version() < 1 {
			return fmt.Errorf("consumer %s: version must be at least 1", id)
		}
		if seen[id] {
			return fmt.Errorf("duplicate consumer %q", id)
		}
		seen[id] = true
		if consumer.New() == nil {
			return fmt.Errorf("consumer %s: New returned nil", id)
		}
	}
	return nil
}

// typeSet is the union of every consumer's watched types.
func typeSet(consumers []Consumer) map[string]bool {
	set := make(map[string]bool)
	for _, consumer := range consumers {
		for _, typ := range consumer.Types() {
			set[typ] = true
		}
	}
	return set
}

// sortedIDs is a stable order for messages and tests.
func sortedIDs(consumers []Consumer) []string {
	ids := make([]string, 0, len(consumers))
	for _, consumer := range consumers {
		ids = append(ids, consumer.ID())
	}
	sort.Strings(ids)
	return ids
}
