package sessionactivity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

type timedDshEvent struct {
	typ  string
	at   int64
	data string
}

// writeDSHTimedLog writes one session log whose rows carry the instants the
// classifier orders a liveness answer against.
func writeDSHTimedLog(t *testing.T, id string, events ...timedDshEvent) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("DSH_HOME", home)
	t.Setenv("HOME", home)
	header, err := json.Marshal(map[string]any{
		"type": "session", "version": 3, "id": id, "createdAt": 10,
		"cwd": "/workspace", "isSeeded": false, "delegationDepth": 0, "agentPreset": "std",
	})
	if err != nil {
		t.Fatal(err)
	}
	body := append(header, '\n')
	for i, event := range events {
		line, err := json.Marshal(map[string]any{
			"type": event.typ, "seq": i, "time": event.at, "data": json.RawMessage(event.data),
		})
		if err != nil {
			t.Fatal(err)
		}
		body = append(append(body, line...), '\n')
	}
	path := filepath.Join(home, "sessions", "p", id, "session.v3.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

// An open turn whose process stopped is an error only when the liveness answer
// is at least as new as the turn. A cached answer that predates the turn must
// leave the log's phase standing: that is the red banner this rule removes.
func TestReadDSHStatusAbstainsOnAnOlderLivenessAnswer(t *testing.T) {
	writeDSHTimedLog(t, "s-stale-answer", timedDshEvent{"turn/start", 1000, `{"turn":1}`})

	stale := RunStatus(func(string) Lookup {
		return Lookup{Present: true, Answered: true, At: time.UnixMilli(999)}
	})
	snap, err := readDSHStatus("s-stale-answer", verdictSources(stale, DurableOutcome{}, false, true))
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Thinking || snap.Live || !snap.LiveStale {
		t.Fatalf("%+v, an answer older than the turn cannot call it dead", snap)
	}
	if snap.Detail != "" || snap.Unanswered {
		t.Fatalf("%+v, an abstained read is not an error and not unanswered", snap)
	}

	fresh := RunStatus(func(string) Lookup {
		return Lookup{Present: true, Answered: true, At: time.UnixMilli(1000)}
	})
	snap, err = readDSHStatus("s-stale-answer", verdictSources(fresh, DurableOutcome{}, false, true))
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != ErrorKind || snap.Detail != DetailTurnOpen || snap.LiveStale {
		t.Fatalf("%+v, a turn that outlives the answer really is dead", snap)
	}
	if snap.Unanswered {
		t.Fatalf("%+v, a real death is answered", snap)
	}
}

// The follow path shares the rule, so the transcript card and the activity grid
// agree about a session whose answer predates its open turn.
func TestFollowDSHStatusAbstainsOnAnOlderLivenessAnswer(t *testing.T) {
	writeDSHTimedLog(t, "s-stale-follow", timedDshEvent{"turn/start", 1000, `{"turn":1}`})
	stale := RunStatus(func(string) Lookup {
		return Lookup{Present: true, Answered: true, At: time.UnixMilli(999)}
	})
	snap, rows, err := followDSHStatus("s-stale-follow", 0, 10, verdictSources(stale, DurableOutcome{}, false, true))
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Thinking || snap.Live || !snap.LiveStale {
		t.Fatalf("%+v rows=%d", snap, len(rows))
	}
}

// The tail fold takes the observation, not a boolean, so the instant it needs
// to order the answer against the log survives the lookup mapping.
func TestLiveFromLookupKeepsTheAnswerInstant(t *testing.T) {
	at := time.UnixMilli(1234)
	cases := []struct {
		name string
		look Lookup
		want dshsessions.Live
	}{
		{
			name: "running",
			look: Lookup{Running: true, Present: true, Answered: true, At: at},
			want: dshsessions.Live{Running: true, Known: true, At: at},
		},
		{
			name: "listed but not running",
			look: Lookup{Present: true, Answered: true, At: at},
			want: dshsessions.Live{Known: true, At: at},
		},
		{
			name: "absent from an answered map",
			look: Lookup{Answered: true, At: at},
			want: dshsessions.Live{At: at},
		},
		{
			name: "unanswered rpc",
			look: Lookup{},
			want: dshsessions.Live{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := liveFromLookup(tc.look); got != tc.want {
				t.Fatalf("got %+v want %+v", got, tc.want)
			}
		})
	}
}
