package cli

import (
	"bytes"
	"encoding/json"
	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpEveryLevel(t *testing.T) {
	for _, args := range [][]string{nil, {"--help"}, {"session"}, {"session", "-h"}, {"sessions", "--help"}, {"session", "list", "--help"}, {"session", "info", "-h"}, {"session", "status", "--help"}, {"session", "log", "--help"}, {"session", "messages", "--help"}, {"session", "prompts", "--help"}} {
		var out, diag bytes.Buffer
		if err := Run(args, &out, &diag); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !strings.Contains(out.String(), "Usage: agent-pro dsh") || diag.Len() != 0 {
			t.Fatalf("%v: %q %q", args, out.String(), diag.String())
		}
	}
}
func TestInvalidCLI(t *testing.T) {
	for _, args := range [][]string{{"send"}, {"session", "send"}, {"sessions", "extra"}, {"sessions", "--limit", "0"}, {"session", "info"}, {"session", "log", "x", "--tail", "-1"}, {"sessions", "--unknown"}} {
		var out, diag bytes.Buffer
		if err := Run(args, &out, &diag); err == nil {
			t.Errorf("%v accepted", args)
		}
	}
}
func TestFilter(t *testing.T) {
	dir := t.TempDir()
	rows := []sessions.Summary{{ID: "z", Cwd: dir, UpdatedAt: 2}, {ID: "a", Cwd: dir, UpdatedAt: 2}, {ID: "outside", Cwd: "/other", UpdatedAt: 9}}
	got, err := filter(rows, false, []string{dir}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("%v %v", got, err)
	}
	if _, err := filter(rows, false, []string{filepath.Join(dir, "missing")}, 20); err == nil {
		t.Fatal("missing directory accepted")
	}
}
func transcript() *sessions.Inspection {
	return &sessions.Inspection{Summary: sessions.Summary{ID: "s", Outcome: "completed"}, Events: []sessions.Event{
		{Type: "user/message", Seq: 0, Data: json.RawMessage(`{"content":[{"type":"text","text":"human"}],"source":{"kind":"user"}}`)},
		{Type: "user/message", Seq: 1, Data: json.RawMessage(`{"content":[{"type":"text","text":"notice"}],"source":{"kind":"plugin"}}`)},
		{Type: "assistant/message", Seq: 2, Data: json.RawMessage(`{"message":{"content":[{"type":"text","text":"answer"}]},"stream":["answer"]}`)},
		{Type: "assistant/attempt", Seq: 3, Data: json.RawMessage(`{"message":{"content":[{"type":"text","text":"failed"}]}}`)},
		{Type: "tool/call", Seq: 4, Data: json.RawMessage(`{"name":"bash","arguments":{"command":"pwd"}}`)},
	}}
}
func TestMessagesAndLog(t *testing.T) {
	in := transcript()
	human, err := messages(in, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(human) != 1 || human[0].Text != "human" {
		t.Fatalf("%+v", human)
	}
	all, err := messages(in, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 || all[2].Text != "answer" {
		t.Fatalf("%+v", all)
	}
	var out bytes.Buffer
	if err := printLog(&out, in, 0, false); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "answer") != 1 || strings.Contains(out.String(), "failed") || !strings.Contains(out.String(), "pwd") {
		t.Fatal(out.String())
	}
	out.Reset()
	if err := printLog(&out, in, 1, true); err != nil {
		t.Fatal(err)
	}
	if strings.Count(strings.TrimSpace(out.String()), "\n") != 0 || !json.Valid(out.Bytes()) {
		t.Fatal(out.String())
	}
}
func TestStatusScope(t *testing.T) {
	var out bytes.Buffer
	if err := printInfo(&out, transcript().Summary, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "process liveness is unknown") {
		t.Fatal(out.String())
	}
}
func TestEmptyRootJSON(t *testing.T) {
	root := filepath.Join(t.TempDir(), "absent")
	var out, diag bytes.Buffer
	if err := Run([]string{"sessions", "--sessions-dir", root, "--json"}, &out, &diag); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "[]" || diag.Len() != 0 {
		t.Fatalf("%q %q", out.String(), diag.String())
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("root modified: %v", err)
	}
}
