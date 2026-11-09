package sessionactivity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type dshEvent struct {
	typ  string
	data string
}

func turnStart() dshEvent {
	return dshEvent{"turn/start", `{"turn":1}`}
}

func turnEnd(reason string) dshEvent {
	return dshEvent{"turn/end", `{"turn":1,"reason":{"kind":"` + reason + `"}}`}
}

// writeDSHLog writes one session.v3.jsonl under a temp DSH_HOME. Every fixture
// uses its own id because dshSessions memoizes tails per id for the life of the
// process.
func writeDSHLog(t *testing.T, id string, events ...dshEvent) {
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
			"type": event.typ, "seq": i, "time": 20 + i, "data": json.RawMessage(event.data),
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

// presentIdle makes session/status answer with a known, not-running session:
// the exact shape eluc gets for a session that already exists but has no live
// process.
func presentIdle(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { SetDefaultRunStatus(nil) })
	SetDefaultRunStatus(func(string) Lookup { return Lookup{Answered: true, Present: true} })
}

func TestNotRunningClassifiesTail(t *testing.T) {
	presentIdle(t)
	// A present-but-idle session is not "unknown": session/status answered, so
	// liveness is known to be false, and the tail still owns the last word on
	// how the last turn ended.
	cases := []struct {
		name       string
		events     []dshEvent
		wantKind   Kind
		wantDetail string
	}{
		{
			name:       "failed turn stays error",
			events:     []dshEvent{turnStart(), turnEnd("error")},
			wantKind:   ErrorKind,
			wantDetail: "error",
		},
		{
			name:       "aborted turn stays aborted",
			events:     []dshEvent{turnStart(), turnEnd("aborted")},
			wantKind:   Aborted,
			wantDetail: "aborted",
		},
		{
			name:     "completed turn is idle",
			events:   []dshEvent{turnStart(), turnEnd("completed")},
			wantKind: Idle,
		},
		{
			name:       "open turn without liveness is an error",
			events:     []dshEvent{turnStart()},
			wantKind:   ErrorKind,
			wantDetail: "turn open but session is not live",
		},
		{
			name:     "empty log is idle",
			wantKind: Idle,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := "s-not-running-" + tc.name
			writeDSHLog(t, id, tc.events...)
			got, err := Read(RunnerDSH, id)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if got.Kind != tc.wantKind {
				t.Fatalf("kind=%s want %s (outcome %q)", got.Kind, tc.wantKind, got.Outcome)
			}
			if got.Detail != tc.wantDetail {
				t.Fatalf("detail=%q want %q", got.Detail, tc.wantDetail)
			}
			if got.Live {
				t.Fatal("a not-running session must never be live")
			}
		})
	}
}

// The failed case is the regression this test exists for: v0.0.229 reported
// kind=idle with outcome=error for exactly this log.
func TestNotRunningKeepsOutcomeAndKindInStep(t *testing.T) {
	presentIdle(t)
	id := "s-not-running-outcome"
	writeDSHLog(t, id, turnStart(), turnEnd("error"))
	got, err := Read(RunnerDSH, id)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Outcome != "error" {
		t.Fatalf("outcome=%q want error", got.Outcome)
	}
	if got.Kind != ErrorKind {
		t.Fatalf("kind=%s contradicts outcome %q", got.Kind, got.Outcome)
	}
}

func TestNotRunningTellsAbsentFromUnreachable(t *testing.T) {
	presentIdle(t)
	id := "s-not-running-presents"
	writeDSHLog(t, id, turnStart(), turnEnd("completed"))

	tail, err := dshTail(id)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		look Lookup
		want Kind
	}{
		{name: "unanswered rpc is error", look: Lookup{}, want: ErrorKind},
		{name: "absent from a successful map is not_found", look: Lookup{Answered: true}, want: NotFound},
		{name: "present and idle classifies the tail", look: Lookup{Answered: true, Present: true}, want: Idle},
		{name: "present and live classifies the tail", look: Lookup{Answered: true, Present: true, Running: true}, want: Idle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := snapshotFromTail(tail, tc.look).Kind; got != tc.want {
				t.Fatalf("kind=%s want %s", got, tc.want)
			}
		})
	}
}
