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
