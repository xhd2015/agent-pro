package sessions

import (
	"encoding/json"
	"testing"
)

func ev(typ string, data string) Event {
	return Event{Type: typ, Data: json.RawMessage(data)}
}

func TestClassifyDSHActivity(t *testing.T) {
	assistantText := `{"message":{"content":[{"type":"text","text":"hello"}]}}`
	assistantThink := `{"message":{"content":[{"type":"reasoning","text":"hmm"}]}}`
	toolCall := `{"name":"bash","arguments":{"command":"pwd"}}`
	turnStart := `{"turn":1}`
	turnDone := `{"turn":1,"reason":{"kind":"completed"}}`
	turnErr := `{"turn":1,"reason":{"kind":"error"}}`
	userMsg := `{"content":[{"type":"text","text":"hi"}],"source":{"kind":"user"}}`

	cases := []struct {
		name     string
		events   []Event
		live     bool
		wantKind ActivityKind
		wantDet  string
	}{
		{name: "empty idle", wantKind: ActivityIdle},
		{
			name:     "open turn thinking",
			events:   []Event{ev("turn/start", turnStart)},
			live:     true,
			wantKind: ActivityThinking,
		},
		{
			name:     "assistant attempt without text is thinking",
			events:   []Event{ev("turn/start", turnStart), ev("assistant/attempt", assistantThink)},
			live:     true,
			wantKind: ActivityThinking,
		},
		{
			name:     "assistant text is responding",
			events:   []Event{ev("turn/start", turnStart), ev("assistant/message", assistantText)},
			live:     true,
			wantKind: ActivityResponding,
		},
		{
			name:     "tool call",
			events:   []Event{ev("turn/start", turnStart), ev("tool/call", toolCall)},
			live:     true,
			wantKind: ActivityToolCalling,
			wantDet:  "bash",
		},
		{
			name:     "ask user question waits",
			events:   []Event{ev("turn/start", turnStart), ev("tool/call", `{"name":"ask_user_question"}`)},
			live:     true,
			wantKind: ActivityWaiting,
			wantDet:  "ask_user_question",
		},
		{
			name:     "dead ask user question stays an error",
			events:   []Event{ev("turn/start", turnStart), ev("tool/call", `{"name":"ask_user_question"}`)},
			live:     false,
			wantKind: ActivityError,
			wantDet:  "ask_user_question",
		},
		{
			name:     "tool result returns to thinking",
			events:   []Event{ev("turn/start", turnStart), ev("tool/call", toolCall), ev("tool/result", `{"name":"bash"}`)},
			live:     true,
			wantKind: ActivityThinking,
		},
		{
			name:     "completed turn idle",
			events:   []Event{ev("turn/start", turnStart), ev("assistant/message", assistantText), ev("turn/end", turnDone)},
			live:     true,
			wantKind: ActivityIdle,
		},
		{
			name:     "turn end error not live",
			events:   []Event{ev("turn/start", turnStart), ev("turn/end", turnErr)},
			live:     false,
			wantKind: ActivityError,
			wantDet:  "error",
		},
		{
			name:     "turn end error while live stays error",
			events:   []Event{ev("turn/start", turnStart), ev("turn/end", turnErr)},
			live:     true,
			wantKind: ActivityError,
			wantDet:  "error",
		},
		{
			name:     "turn end aborted while live",
			events:   []Event{ev("turn/start", turnStart), ev("turn/end", `{"turn":1,"reason":{"kind":"aborted"}}`)},
			live:     true,
			wantKind: ActivityAborted,
			wantDet:  "aborted",
		},
		{
			name:     "turn end aborted not live",
			events:   []Event{ev("turn/start", turnStart), ev("turn/end", `{"turn":1,"reason":{"kind":"aborted"}}`)},
			live:     false,
			wantKind: ActivityAborted,
			wantDet:  "aborted",
		},
		{
			name:     "open turn not live is error",
			events:   []Event{ev("turn/start", turnStart), ev("assistant/message", assistantText)},
			live:     false,
			wantKind: ActivityError,
			wantDet:  "turn open but session is not live",
		},
		{
			name:     "user message while live is thinking",
			events:   []Event{ev("user/message", userMsg)},
			live:     true,
			wantKind: ActivityThinking,
		},
		{
			name:     "user message while not live is idle",
			events:   []Event{ev("user/message", userMsg)},
			live:     false,
			wantKind: ActivityIdle,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.events, tc.live)
			if got.Kind != tc.wantKind || got.Detail != tc.wantDet || got.Live != tc.live {
				t.Fatalf("got %+v want kind=%s detail=%q live=%v", got, tc.wantKind, tc.wantDet, tc.live)
			}
		})
	}
}

func TestClassifyLogSkipsNotLiveOverride(t *testing.T) {
	got := ClassifyLog([]Event{ev("turn/start", `{"turn":1}`)})
	if got.Kind != ActivityThinking || got.Live {
		t.Fatalf("%+v", got)
	}
}
