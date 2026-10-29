package sessionactivity

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestFollowDSHUnchangedReusesFile(t *testing.T) {
	home := t.TempDir()
	id := "s-follow-cache"
	header, _ := json.Marshal(map[string]any{
		"type": "session", "version": 3, "id": id, "createdAt": 10,
		"cwd": "/workspace", "isSeeded": false, "delegationDepth": 0, "agentPreset": "std",
	})
	row, _ := json.Marshal(map[string]any{
		"type": "tool/call", "seq": 0, "time": 20,
		"data": json.RawMessage(`{"name":"read","arguments":{"path":"f"}}`),
	})
	path := filepath.Join(home, "sessions", "p", id, "session.v3.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append(header, '\n'), append(row, '\n')...), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DSH_HOME", home)
	t.Setenv("HOME", home)
	first, rows, err := Follow(RunnerDSH, id, 0, 24)
	if err != nil || len(rows) != 1 {
		t.Fatalf("first %+v %v %v", first, rows, err)
	}
	second, rows2, err := Follow(RunnerDSH, id, 0, 24)
	if err != nil || second.Kind != first.Kind || len(rows2) != 1 || rows2[0].Tool != rows[0].Tool {
		t.Fatalf("second %+v %v %v", second, rows2, err)
	}
}
