package sessions

import (
	"encoding/json"
	"reflect"
	"testing"
)

// contextEvent is one injected user/message, written the way DSH v4 logs it.
func contextEvent(seq int64, source, text string) Event {
	data := `{"source":` + source + `,"content":[{"type":"text","text":` + quoteJSON(text) + `}]}`
	return Event{Type: "user/message", Seq: seq, Time: seq, Data: json.RawMessage(data)}
}

func quoteJSON(value string) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}

func contextRows(messages []Message) []Message {
	out := make([]Message, 0, len(messages))
	for _, message := range messages {
		if message.Kind == TranscriptContext {
			out = append(out, message)
		}
	}
	return out
}

func TestProjectContextInjectionLabels(t *testing.T) {
	events := []Event{
		contextEvent(1, `{"kind":"skill-invocation","name":"brainstorm","form":"instructions"}`, "<skill_content name=\"brainstorm\">"),
		contextEvent(2, `{"kind":"skill-catalog","form":"catalog"}`, "A skill is a reusable set of instructions."),
		contextEvent(3, `{"kind":"agent-instructions","form":"instructions","changes":[{"action":"set","path":"AGENTS.md"},{"action":"set","path":"AGENTS.md"},{"action":"set","path":"CLAUDE.md"}]}`, "<system-reminder>workspace</system-reminder>"),
		contextEvent(4, `{"kind":"runtime-context","form":"snapshot"}`, "Current working directory: /tmp/x"),
		contextEvent(5, `{"kind":"tool-jobs","form":"notice","summary":"bash tests finished"}`, "background job bash-1 finished"),
		contextEvent(6, `{"kind":"session-reference","form":"recall","references":[{"label":"messages-parity"},{"label":"dsh-audit"},{"label":"messages-parity"}]}`, "recalled context"),
	}
	rows := contextRows(Project(events))
	if len(rows) != len(events) {
		t.Fatalf("rows=%d want %d: %+v", len(rows), len(events), rows)
	}
	want := []struct {
		source, form, label, summary string
	}{
		{"skill-invocation", "instructions", "brainstorm", ""},
		{"skill-catalog", "catalog", "skill-catalog", ""},
		{"agent-instructions", "instructions", "AGENTS.md, CLAUDE.md", ""},
		{"runtime-context", "snapshot", "runtime-context", ""},
		{"tool-jobs", "notice", "tool-jobs", "bash tests finished"},
		{"session-reference", "recall", "messages-parity, dsh-audit", ""},
	}
	for i, expected := range want {
		row := rows[i]
		if row.Injection == nil {
			t.Fatalf("row[%d] has no injection: %+v", i, row)
		}
		got := *row.Injection
		if got.Source != expected.source || got.Form != expected.form || got.Label != expected.label || got.Summary != expected.summary {
			t.Fatalf("row[%d]=%+v want %+v", i, got, expected)
		}
	}
	if rows[0].Text != "<skill_content name=\"brainstorm\">" {
		t.Fatalf("text=%q", rows[0].Text)
	}
	if rows[3].Seq != 4 || rows[3].Time != 4 {
		t.Fatalf("seq/time=%d/%d", rows[3].Seq, rows[3].Time)
	}
}

func TestProjectContextSkipsHumanAndCompactionCheckpoint(t *testing.T) {
	for _, source := range []string{
		`{"kind":"user"}`,
		`{"kind":"compact-checkpoint","compactionId":"k1"}`,
	} {
		event := contextEvent(1, source, "hello")
		if injection, ok := messageInjection(event); ok {
			t.Fatalf("source %s projected as %+v", source, injection)
		}
	}

	// A missing source is the one case that is neither: an older or synthetic
	// log still gets a context row, labelled by its empty kind, instead of
	// vanishing from the transcript.
	event := Event{Type: "user/message", Seq: 9, Time: 9, Data: json.RawMessage(`{"content":[{"type":"text","text":"notice"}]}`)}
	injection, ok := messageInjection(event)
	if !ok || injection.Source != "" || injection.Label != "" {
		t.Fatalf("missing source: %+v ok=%v", injection, ok)
	}
	rows := Project([]Event{event})
	if len(rows) != 1 || rows[0].Kind != TranscriptContext {
		t.Fatalf("rows=%+v", rows)
	}
}

// The Messages tab reads the cached cursor, not Project: context rows must
// reach it the same way, or the pane that polls shows a different transcript
// from the one a cold read builds.
func TestViewTailContextRowsMatchProject(t *testing.T) {
	events := []Event{
		{Type: "turn/start", Seq: 1, Time: 1, Data: json.RawMessage(`{"turn":1}`)},
		contextEvent(2, `{"kind":"skill-catalog","form":"catalog"}`, "catalog"),
		contextEvent(3, `{"kind":"skill-invocation","name":"brainstorm","form":"instructions"}`, "skill"),
		{Type: "assistant/message", Seq: 4, Time: 4, Data: json.RawMessage(`{"message":{"content":[{"type":"text","text":"ok"}]}}`)},
	}
	cursor := &viewTailCursor{}
	for _, event := range events {
		cursor.applyEvent(event)
	}
	want := Project(events)
	if len(cursor.rows) != len(want) {
		t.Fatalf("cursor=%d project=%d", len(cursor.rows), len(want))
	}
	for i := range want {
		if !reflect.DeepEqual(cursor.rows[i], want[i]) {
			t.Fatalf("row[%d] cursor=%+v project=%+v", i, cursor.rows[i], want[i])
		}
	}
}
