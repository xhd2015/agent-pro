package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "project", "s")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := `{"type":"session","version":3,"id":"s","createdAt":10,"cwd":"/workspace","isSeeded":false,"delegationDepth":0}` + "\n"
	for _, event := range transcript().Events {
		raw, err := json.Marshal(map[string]any{"type": event.Type, "seq": event.Seq, "time": event.Seq + 20, "data": event.Data})
		if err != nil {
			t.Fatal(err)
		}
		body += string(raw) + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "session.v3.jsonl"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}
func TestRecentPromptsTailUsesChronology(t *testing.T) {
	root := fixtureRoot(t)
	dir := filepath.Join(root, "project", "newer")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := `{"type":"session","version":3,"id":"newer","createdAt":30,"cwd":"/workspace","isSeeded":false,"delegationDepth":0}` + "\n" + `{"type":"user/message","seq":0,"time":40,"data":{"content":[{"type":"text","text":"latest"}],"source":{"kind":"user"}}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "session.v3.jsonl"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		flags []string
		want  string
	}{{nil, "human,latest"}, {[]string{"--tail", "1"}, "latest"}, {[]string{"--limit", "1"}, "latest"}} {
		var out, diag bytes.Buffer
		args := append([]string{"session", "prompts", "--sessions-dir", root, "--json"}, tc.flags...)
		if err := Run(args, &out, &diag); err != nil {
			t.Fatal(err)
		}
		var items []Message
		if err := json.Unmarshal(out.Bytes(), &items); err != nil {
			t.Fatal(err)
		}
		texts := []string{}
		for _, item := range items {
			texts = append(texts, item.Text)
		}
		if strings.Join(texts, ",") != tc.want {
			t.Fatalf("%v: %s", tc.flags, out.String())
		}
	}
}

func TestCommandsReadFixture(t *testing.T) {
	root := fixtureRoot(t)
	for _, action := range []string{"info", "status", "messages", "prompts", "log"} {
		t.Run(action, func(t *testing.T) {
			var out, diag bytes.Buffer
			if err := Run([]string{"session", action, "s", "--sessions-dir", root, "--json"}, &out, &diag); err != nil {
				t.Fatal(err)
			}
			if diag.Len() != 0 {
				t.Fatal(diag.String())
			}
			if action == "log" {
				for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
					if !json.Valid([]byte(line)) {
						t.Fatal(line)
					}
				}
			} else if !json.Valid(out.Bytes()) {
				t.Fatal(out.String())
			}
			if action == "prompts" && (!strings.Contains(out.String(), "human") || strings.Contains(out.String(), "notice")) {
				t.Fatal(out.String())
			}
		})
	}
	var first, alias, diag bytes.Buffer
	if err := Run([]string{"sessions", "--sessions-dir", root, "--json"}, &first, &diag); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"session", "list", "--sessions-dir", root, "--json"}, &alias, &diag); err != nil {
		t.Fatal(err)
	}
	if first.String() != alias.String() {
		t.Fatal("alias differs")
	}
}
func TestWarningsKeepJSONClean(t *testing.T) {
	root := fixtureRoot(t)
	dir := filepath.Join(root, "project", "broken")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "session.v3.jsonl"), []byte("broken\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if err := Run([]string{"sessions", "--sessions-dir", root, "--json"}, &out, &diag); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(out.Bytes()) || !strings.HasPrefix(diag.String(), "warning:") {
		t.Fatalf("%q %q", out.String(), diag.String())
	}
}
func TestPromptOutputSnapshot(t *testing.T) {
	root := fixtureRoot(t)
	var out, diag bytes.Buffer
	if err := Run([]string{"session", "prompts", "s", "--sessions-dir", root}, &out, &diag); err != nil {
		t.Fatal(err)
	}
	const want = "[1970-01-01T00:00:00Z] user s (#0)\nhuman\n\n"
	if out.String() != want || diag.String() != "" {
		t.Fatalf("stdout = %q; stderr = %q", out.String(), diag.String())
	}
}

func TestStatusJSONIncludesActivity(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "project", "done")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := `{"type":"session","version":3,"id":"done","createdAt":10,"cwd":"/workspace","isSeeded":false,"delegationDepth":0}` + "\n" +
		`{"type":"turn/start","seq":0,"time":20,"data":{"turn":1}}` + "\n" +
		`{"type":"assistant/message","seq":1,"time":21,"data":{"message":{"content":[{"type":"text","text":"ok"}]}}}` + "\n" +
		`{"type":"turn/end","seq":2,"time":22,"data":{"turn":1,"reason":{"kind":"completed"}}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "session.v3.jsonl"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if err := Run([]string{"session", "status", "done", "--sessions-dir", root, "--json"}, &out, &diag); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["activity"] != "idle" {
		t.Fatalf("activity=%v body=%s", got["activity"], out.String())
	}
	if _, ok := got["live"]; !ok {
		t.Fatal("missing live")
	}
}

func TestRejectIgnoredOptions(t *testing.T) {
	for _, args := range [][]string{{"session", "info", "s", "--tail", "2"}, {"session", "status", "s", "--tail", "2"}, {"session", "prompts", "s", "--limit", "20"}, {"session", "prompts", "s", "--here"}} {
		var out, diag bytes.Buffer
		if err := Run(args, &out, &diag); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}
