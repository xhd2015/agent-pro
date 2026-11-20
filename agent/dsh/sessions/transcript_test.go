package sessions

import (
	"encoding/json"
	"testing"
)

func TestTranscriptUserAssistantTool(t *testing.T) {
	events := []Event{
		ev("user/message", `{"content":[{"type":"text","text":"hello"}],"source":{"kind":"user"}}`),
		ev("user/message", `{"content":[{"type":"text","text":"plugin"}],"source":{"kind":"plugin"}}`),
		ev("assistant/message", `{"message":{"content":[{"type":"text","text":"hi"}]}}`),
		ev("tool/call", `{"name":"bash","arguments":{"command":"pwd"}}`),
	}
	// Fix seqs
	for i := range events {
		events[i].Seq = int64(i)
		events[i].Time = int64(20 + i)
	}
	got := Transcript(events, 0, 24)
	if len(got) != 3 {
		t.Fatalf("%+v", got)
	}
	if got[0].Kind != TranscriptUser || got[0].Text != "hello" {
		t.Fatalf("user=%+v", got[0])
	}
	if got[1].Kind != TranscriptAssistant || got[1].Text != "hi" {
		t.Fatalf("assistant=%+v", got[1])
	}
	if got[2].Kind != TranscriptTool || got[2].Tool != "bash" || got[2].Text != "pwd" {
		t.Fatalf("tool=%+v", got[2])
	}
}

func TestTranscriptAfterSeqAndTail(t *testing.T) {
	events := make([]Event, 0, 6)
	for i, text := range []string{"a", "b", "c", "d"} {
		events = append(events, Event{
			Type: "user/message",
			Seq:  int64(i + 1),
			Time: int64(i),
			Data: json.RawMessage(`{"content":[{"type":"text","text":"` + text + `"}],"source":{"kind":"user"}}`),
		})
	}
	tail := Transcript(events, 0, 2)
	if len(tail) != 2 || tail[0].Text != "c" || tail[1].Text != "d" {
		t.Fatalf("tail=%+v", tail)
	}
	later := Transcript(events, 2, 24)
	if len(later) != 2 || later[0].Text != "c" || later[1].Seq != 4 {
		t.Fatalf("after=%+v", later)
	}
}

// A message/edit arrives after the message it changes, so the transcript cannot
// be a per-event mapping: the fold has to apply the edit to a row it already
// produced.
func TestTranscriptAppliesMessageEdits(t *testing.T) {
	events := []Event{
		ev("user/message", `{"content":[{"type":"text","text":"original question"}],"source":{"kind":"user"}}`),
		ev("assistant/message", `{"message":{"content":[{"type":"text","text":"original answer"}]}}`),
	}
	for i := range events {
		events[i].Seq = int64(i)
		events[i].Time = int64(20 + i)
	}
	events = append(events, ev("message/edit", `{"edits":[{"seq":1,"text":"rewritten answer"}]}`))
	events[2].Seq = 2

	got := Transcript(events, 0, 24)
	if len(got) != 2 {
		t.Fatalf("rows = %d, want 2: %+v", len(got), got)
	}
	// The row keeps the seq of the message it came from.
	if got[1].Seq != 1 || got[1].Text != "rewritten answer" {
		t.Fatalf("edited row = %+v, want seq 1 with the new text", got[1])
	}
	if got[0].Text != "original question" {
		t.Fatalf("an unrelated row changed: %+v", got[0])
	}
}

// A removal drops the row from the model-visible transcript but the fold must not
// renumber anything: every other seq is what it was.
func TestTranscriptAppliesMessageRemoval(t *testing.T) {
	events := []Event{
		ev("user/message", `{"content":[{"type":"text","text":"q"}],"source":{"kind":"user"}}`),
		ev("assistant/message", `{"message":{"content":[{"type":"text","text":"a"}]}}`),
		ev("assistant/message", `{"message":{"content":[{"type":"text","text":"b"}]}}`),
	}
	for i := range events {
		events[i].Seq = int64(i)
	}
	events = append(events, ev("message/edit", `{"edits":[{"seq":1,"deleted":true}]}`))
	events[3].Seq = 3

	got := Transcript(events, 0, 24)
	if len(got) != 2 {
		t.Fatalf("rows = %d, want the middle one removed: %+v", len(got), got)
	}
	if got[0].Seq != 0 || got[1].Seq != 2 {
		t.Fatalf("seqs = %d,%d; want 0 and 2 untouched", got[0].Seq, got[1].Seq)
	}
	if got[1].Text != "b" {
		t.Fatalf("kept the wrong row: %+v", got[1])
	}
}

// An edit whose target a compaction already shadowed is skipped, not fatal: the
// edit was valid when written, and a committed log must stay readable.
func TestTranscriptSkipsEditOfAbsentTarget(t *testing.T) {
	events := []Event{
		ev("user/message", `{"content":[{"type":"text","text":"q"}],"source":{"kind":"user"}}`),
		ev("message/edit", `{"edits":[{"seq":99,"text":"nowhere"}]}`),
	}
	events[0].Seq = 0
	events[1].Seq = 1

	got := Transcript(events, 0, 24)
	if len(got) != 1 || got[0].Text != "q" {
		t.Fatalf("rows = %+v, want the untouched user row", got)
	}
}

// Two edits of one seq apply in log order: the later one wins.
func TestTranscriptLaterEditWins(t *testing.T) {
	events := []Event{
		ev("assistant/message", `{"message":{"content":[{"type":"text","text":"v0"}]}}`),
	}
	events[0].Seq = 0
	events = append(events,
		ev("message/edit", `{"edits":[{"seq":0,"text":"v1"}]}`),
		ev("message/edit", `{"edits":[{"seq":0,"text":"v2"}]}`))
	events[1].Seq = 1
	events[2].Seq = 2

	got := Transcript(events, 0, 24)
	if len(got) != 1 || got[0].Text != "v2" {
		t.Fatalf("rows = %+v, want v2", got)
	}
}
