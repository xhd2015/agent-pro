package sessions

import (
	"testing"
	"time"
)

// timedEv is an event stamped with the log instant the classifier orders
// liveness against.
func timedEv(typ string, at int64, data string) Event {
	return Event{Type: typ, Data: []byte(data), Time: at}
}

// An answer sampled before the turn opened cannot testify about it: the log
// says the turn is running, so the session is thinking rather than failed. This
// is the false alarm the timestamps exist to remove.
func TestClassifyAbstainsWhenTheAnswerPredatesTheOpenTurn(t *testing.T) {
	events := []Event{
		timedEv("turn/start", 1000, `{"turn":1}`),
	}
	got := Classify(events, Live{Known: true, At: time.UnixMilli(999)})
	if got.Kind != ActivityThinking || got.Detail != "" {
		t.Fatalf("%+v, an older answer must not blame the turn", got)
	}
	if got.Live {
		t.Fatalf("%+v, the session is not live", got)
	}
	if !got.LiveStale {
		t.Fatalf("%+v, an abstained read must say so", got)
	}
}

// An answer sampled at the very instant the turn opened is simultaneous with
// it, so it can testify: the same session really is gone.
func TestClassifyKeepsTheErrorWhenTheAnswerReachesTheOpenTurn(t *testing.T) {
	events := []Event{timedEv("turn/start", 1000, `{"turn":1}`)}
	for _, at := range []int64{1000, 2000} {
		got := Classify(events, Live{Known: true, At: time.UnixMilli(at)})
		if got.Kind != ActivityError || got.Detail != "turn open but session is not live" {
			t.Fatalf("at=%d: %+v, a turn that outlives its answer is still an error", at, got)
		}
		if got.LiveStale {
			t.Fatalf("at=%d: %+v, the answer testified", at, got)
		}
	}
}

// A log or an answer without an instant keeps the historical verdict, so
// fixtures and events written before the timestamps existed behave as before.
func TestClassifyWithoutInstantsKeepsTheHistoricalError(t *testing.T) {
	unstamped := []Event{ev("turn/start", `{"turn":1}`)}
	stamped := []Event{timedEv("turn/start", 1000, `{"turn":1}`)}
	for _, tc := range []struct {
		name   string
		events []Event
		live   Live
	}{
		{name: "unstamped event", events: unstamped, live: Live{Known: true, At: time.UnixMilli(5000)}},
		{name: "unstamped answer", events: stamped, live: Live{Known: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.events, tc.live)
			if got.Kind != ActivityError || got.LiveStale {
				t.Fatalf("%+v, no instant on one side keeps the old verdict", got)
			}
		})
	}
}

// openedAt keeps the earliest event that opened the turn, so a fold that starts
// mid-turn can only make an old answer abstain more often, never blame the turn.
func TestClassifyUsesTheEarliestOpeningEvent(t *testing.T) {
	events := []Event{
		timedEv("assistant/message", 900, `{"message":{"content":[{"type":"text","text":"hi"}]}}`),
		timedEv("turn/start", 1000, `{"turn":1}`),
	}
	if got := Classify(events, Live{Known: true, At: time.UnixMilli(899)}); got.Kind != ActivityThinking || !got.LiveStale {
		t.Fatalf("%+v, the earliest opening event is the turn's start", got)
	}
	if got := Classify(events, Live{Known: true, At: time.UnixMilli(900)}); got.Kind != ActivityError || got.LiveStale {
		t.Fatalf("%+v, an answer at the earliest opening instant testifies", got)
	}
}

// A turn that ended and a new one that started after the answer is the same
// false alarm one turn later: only the newest turn's start matters.
func TestClassifyResetsTheOpenInstantOnTurnEnd(t *testing.T) {
	events := []Event{
		timedEv("turn/start", 1000, `{"turn":1}`),
		timedEv("turn/end", 1100, `{"turn":1,"reason":{"kind":"completed"}}`),
		timedEv("turn/start", 1200, `{"turn":2}`),
	}
	got := Classify(events, Live{Known: true, At: time.UnixMilli(1100)})
	if got.Kind != ActivityThinking || !got.LiveStale {
		t.Fatalf("%+v, turn 2 started after the answer", got)
	}
	got = Classify(events, Live{Known: true, At: time.UnixMilli(1200)})
	if got.Kind != ActivityError || got.LiveStale {
		t.Fatalf("%+v, an answer at turn 2's start testifies", got)
	}
}

// A durable end verdict never depends on liveness, so an old answer cannot
// change how a finished turn is reported.
func TestClassifyDurableEndKindsIgnoreTheAnswerInstant(t *testing.T) {
	for _, reason := range []string{"aborted", "error"} {
		events := []Event{
			timedEv("turn/start", 1000, `{"turn":1}`),
			timedEv("turn/end", 1100, `{"turn":1,"reason":{"kind":"`+reason+`"}}`),
		}
		got := Classify(events, Live{Known: true, At: time.UnixMilli(500)})
		want := ActivityAborted
		if reason == "error" {
			want = ActivityError
		}
		if got.Kind != want || got.Detail != reason || got.LiveStale {
			t.Fatalf("%+v want %s, a finished turn is not a liveness question", got, want)
		}
	}
}

// ClassifyLog never probes liveness, so it must not invent an error or a stale
// read from a missing observation.
func TestClassifyLogNeverErrorsOnLiveness(t *testing.T) {
	got := ClassifyLog([]Event{timedEv("turn/start", 1000, `{"turn":1}`)})
	if got.Kind != ActivityThinking || got.Live || got.LiveStale {
		t.Fatalf("%+v", got)
	}
}

// Tail.Activity is the cursor form of the same contract: it hands the fold the
// observation rather than a bare boolean, so the instant survives the call.
func TestTailActivityCarriesTheObservation(t *testing.T) {
	tail := &Tail{fold: classifyState{
		kind:     ActivityThinking,
		turnOpen: true,
		openedAt: 1000,
	}}
	stale := tail.Activity(Live{Known: true, At: time.UnixMilli(999)})
	if stale.Kind != ActivityThinking || stale.Live || !stale.LiveStale {
		t.Fatalf("%+v", stale)
	}
	fresh := tail.Activity(Live{Known: true, At: time.UnixMilli(1000)})
	if fresh.Kind != ActivityError || fresh.Detail != "turn open but session is not live" {
		t.Fatalf("%+v", fresh)
	}
}
