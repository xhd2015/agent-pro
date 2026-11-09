package consumers

import "testing"

// selectionRows is the pair a live session produces: the pick the user made,
// then the header of the first request that used it.
const (
	selectionNoEffort = `{"provider":"p","model":"m"}`
	selectionHigh     = `{"provider":"p","model":"m","reasoningEffort":"high"}`
	headerHigh        = `{"header":{"config":{"provider":"p","model":"m","reasoningEffort":"high"}}}`
	headerOther       = `{"header":{"config":{"provider":"q","model":"m","reasoningEffort":"high"}}}`
)

func modelOf(t *testing.T, f *logFixture, body string) Snapshot {
	t.Helper()
	f.plain("session-model", body)
	bus := f.bus(Options{Consumers: []Consumer{Model()}})
	snapshot, err := bus.Sync("session-model")
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !snapshot.CaughtUp {
		t.Fatalf("not caught up: %+v", snapshot)
	}
	return snapshot
}

func TestModelFoldClearsPendingWhenTheHeaderConfirmsIt(t *testing.T) {
	f := newFixture(t)
	snapshot := modelOf(t, f, headerAt("session-model", 3)+
		row("model/selection", 1, selectionHigh)+
		row("request/header", 2, headerHigh))

	state := snapshot.Model()
	if state.Pending != nil {
		t.Fatalf("pending = %+v, want nil once the header used it", state.Pending)
	}
	last := state.LastUsed
	if last == nil {
		t.Fatal("lastUsed is nil: the header did not record a route")
	}
	if last.Provider != "p" || last.Model != "m" || last.ReasoningEffort != "high" {
		t.Fatalf("lastUsed = %+v", last)
	}
	if state.Seq != 2 {
		t.Fatalf("seq = %d, want the header's row", state.Seq)
	}
	if next := state.Next(); next == nil || *next != *last {
		t.Fatalf("next = %+v, want the used route", next)
	}
}

// A row-level quirk of the runtime it mirrors: a selection that names no effort
// never matches the header, because the adapter fills the effort in. The next
// request's route is still the header's, so the pending pick stays visible as a
// difference instead of being silently cleared.
func TestModelFoldKeepsPendingWhenTheHeaderResolvedADifferentEffort(t *testing.T) {
	f := newFixture(t)
	snapshot := modelOf(t, f, headerAt("session-model", 3)+
		row("model/selection", 1, selectionNoEffort)+
		row("request/header", 2, headerHigh))

	state := snapshot.Model()
	if state.Pending == nil || state.Pending.Model != "m" || state.Pending.ReasoningEffort != "" {
		t.Fatalf("pending = %+v, want the picked route", state.Pending)
	}
	if state.LastUsed == nil || state.LastUsed.ReasoningEffort != "high" {
		t.Fatalf("lastUsed = %+v, want the served route", state.LastUsed)
	}
	if next := state.Next(); next == nil || next.ReasoningEffort != "" {
		t.Fatalf("next = %+v, want the pending pick", next)
	}
}

func TestModelFoldKeepsPendingWhenTheHeaderServesAnotherRoute(t *testing.T) {
	f := newFixture(t)
	snapshot := modelOf(t, f, headerAt("session-model", 3)+
		row("model/selection", 1, selectionHigh)+
		row("request/header", 2, headerOther))

	state := snapshot.Model()
	if state.Pending == nil || state.Pending.Provider != "p" {
		t.Fatalf("pending = %+v, want the pick the header did not serve", state.Pending)
	}
	if state.LastUsed == nil || state.LastUsed.Provider != "q" {
		t.Fatalf("lastUsed = %+v", state.LastUsed)
	}
}

func TestModelFoldIgnoresARepeatedRoute(t *testing.T) {
	f := newFixture(t)
	snapshot := modelOf(t, f, headerAt("session-model", 3)+
		row("request/header", 2, headerHigh)+
		row("request/header", 3, headerHigh)+
		row("model/selection", 4, `{"provider":"q","model":"z","reasoningEffort":"low"}`))

	state := snapshot.Model()
	if state.LastUsed == nil || state.LastUsed.Provider != "p" {
		t.Fatalf("lastUsed = %+v, want the repeated route kept as it was", state.LastUsed)
	}
	if state.Pending == nil || state.Pending.Model != "z" {
		t.Fatalf("pending = %+v, want the newer pick", state.Pending)
	}
	if state.Seq != 4 {
		t.Fatalf("seq = %d, want the newest row that changed the route", state.Seq)
	}
}

func TestModelFoldSkipsRowsWithoutARoute(t *testing.T) {
	f := newFixture(t)
	snapshot := modelOf(t, f, headerAt("session-model", 3)+
		row("model/selection", 1, `{"provider":"","model":"m"}`)+
		row("request/header", 2, `{"header":{"config":{"tools":[]}}}`))

	state := snapshot.Model()
	if state.LastUsed != nil || state.Pending != nil || state.Seq != 0 {
		t.Fatalf("state = %+v, want an untouched fold", state)
	}
	if next := state.Next(); next != nil {
		t.Fatalf("next = %+v, want no route at all", next)
	}
}
