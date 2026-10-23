package sessionactivity

import (
	"testing"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

func TestChatTranscriptMapsKinds(t *testing.T) {
	rows := chatTranscript([]chatLine{
		{Kind: "thinking", Text: "secret"},
		{Kind: "user", Text: "hello", Time: 1},
		{Kind: "tool", Text: "pwd", Tool: "bash", Time: 2},
		{Kind: "response", Text: "hi", Time: 3},
	}, 0, 24)
	if len(rows) != 3 {
		t.Fatalf("%+v", rows)
	}
	if rows[0].Kind != dshsessions.TranscriptUser || rows[0].Seq != 1 {
		t.Fatalf("%+v", rows[0])
	}
	if rows[1].Kind != dshsessions.TranscriptTool || rows[1].Tool != "bash" {
		t.Fatalf("%+v", rows[1])
	}
	if rows[2].Kind != dshsessions.TranscriptAssistant {
		t.Fatalf("%+v", rows[2])
	}
	later := chatTranscript([]chatLine{
		{Kind: "user", Text: "a"},
		{Kind: "response", Text: "b"},
	}, 1, 24)
	if len(later) != 1 || later[0].Text != "b" {
		t.Fatalf("%+v", later)
	}
}
