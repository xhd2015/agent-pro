package sessions

import (
	"encoding/json"
	"testing"
)

// TestClassifyPausedTurn pins the durable shape a pause writes: an aborted turn
// whose hook cause is "paused". Readers must see paused, never aborted.
func TestClassifyPausedTurn(t *testing.T) {
	events := []Event{
		{Type: "turn/start", Time: 1, Data: json.RawMessage(`{"turn":1}`)},
		{Type: "turn/end", Time: 2, Data: json.RawMessage(`{"turn":1,"reason":{"kind":"aborted","reason":{"kind":"hook","reason":"paused"}}}`)},
	}
	act := Classify(events, Live{Known: true, Running: false})
	if act.Kind != ActivityPaused || act.Detail != PauseReason {
		t.Fatalf("paused turn = %+v", act)
	}
}

// TestProjectPausedBecomesResumed pins one card per pause: a later turn/start
// stamps the same row instead of adding a second record.
func TestProjectPausedBecomesResumed(t *testing.T) {
	rows := Project([]Event{
		{Type: "turn/start", Time: 1, Data: json.RawMessage(`{"turn":1}`)},
		{Type: "turn/end", Time: 2, Data: json.RawMessage(`{"turn":1,"reason":{"kind":"aborted","reason":{"kind":"hook","reason":"paused"}}}`)},
		{Type: "turn/start", Time: 9, Data: json.RawMessage(`{"turn":2}`)},
	})
	var paused []Message
	for _, row := range rows {
		if row.Kind == "paused" || row.Kind == "resumed" {
			paused = append(paused, row)
		}
	}
	if len(paused) != 1 {
		t.Fatalf("pause rows = %+v", paused)
	}
	if paused[0].Kind != "resumed" || paused[0].ResumedAt != 9 {
		t.Fatalf("resumed row = %+v", paused[0])
	}
}
