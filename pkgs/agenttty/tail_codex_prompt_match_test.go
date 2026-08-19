package agenttty

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeRollout(t *testing.T, dir, name, sessionID, cwd, userPrompt string, sessionTime time.Time) string {
	t.Helper()
	path := filepath.Join(dir, name)
	ts := sessionTime.UTC().Format(time.RFC3339Nano)
	// Minimal session_meta + env context user + real user prompt.
	body := `{"timestamp":"` + ts + `","type":"session_meta","payload":{"id":"` + sessionID + `","session_id":"` + sessionID + `","cwd":"` + cwd + `","timestamp":"` + ts + `"}}
{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"<environment_context>\n  <cwd>` + cwd + `</cwd>\n</environment_context>"}]}}
{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"` + userPrompt + `"}]}}
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write rollout: %v", err)
	}
	return path
}

func TestScanActiveCodexTranscripts_PromptMatchDistinct(t *testing.T) {
	root := t.TempDir()
	// Glob pattern: sessions/*/*/*/rollout-*.jsonl
	day := filepath.Join(root, "sessions", "2026", "08", "12")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(root, "ws")
	if err := os.MkdirAll(ws, 0755); err != nil {
		t.Fatal(err)
	}
	absWS, _ := filepath.Abs(ws)
	t0 := time.Now().Add(-30 * time.Second)
	t1 := t0.Add(2 * time.Second)
	writeRollout(t, day, "rollout-a-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.jsonl",
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", absWS, "QUESTION_A", t0)
	writeRollout(t, day, "rollout-b-bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.jsonl",
		"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", absWS, "QUESTION_B", t1)

	runStart := t0.Add(-1 * time.Second)

	idA, _, okA, err := scanActiveCodexTranscripts(root, absWS, runStart, "QUESTION_A")
	if err != nil || !okA || idA != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("prompt A: ok=%v id=%q err=%v want aaaa…", okA, idA, err)
	}
	idB, _, okB, err := scanActiveCodexTranscripts(root, absWS, runStart, "QUESTION_B")
	if err != nil || !okB || idB != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Fatalf("prompt B: ok=%v id=%q err=%v want bbbb…", okB, idB, err)
	}
}

func TestScanActiveCodexTranscripts_MultiMatchFailClosed(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "sessions", "2026", "08", "12")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(root, "ws")
	if err := os.MkdirAll(ws, 0755); err != nil {
		t.Fatal(err)
	}
	absWS, _ := filepath.Abs(ws)
	t0 := time.Now().Add(-30 * time.Second)
	writeRollout(t, day, "rollout-a1-aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.jsonl",
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", absWS, "SAME_PROMPT", t0)
	writeRollout(t, day, "rollout-a2-bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb.jsonl",
		"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", absWS, "SAME_PROMPT", t0.Add(time.Second))

	// Empty prompt + multi cwd matches → fail closed (not newest).
	id, _, ok, err := scanActiveCodexTranscripts(root, absWS, t0.Add(-time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	if ok || id != "" {
		t.Fatalf("empty prompt multi: want fail-closed, got ok=%v id=%q", ok, id)
	}
	// Same prompt multi → fail closed.
	id, _, ok, err = scanActiveCodexTranscripts(root, absWS, t0.Add(-time.Second), "SAME_PROMPT")
	if err != nil {
		t.Fatal(err)
	}
	if ok || id != "" {
		t.Fatalf("same prompt multi: want fail-closed, got ok=%v id=%q", ok, id)
	}
}

func writeRolloutUsers(t *testing.T, dir, name, sessionID, cwd string, users []string, sessionTime time.Time) string {
	t.Helper()
	path := filepath.Join(dir, name)
	ts := sessionTime.UTC().Format(time.RFC3339Nano)
	body := `{"timestamp":"` + ts + `","type":"session_meta","payload":{"id":"` + sessionID + `","session_id":"` + sessionID + `","cwd":"` + cwd + `","timestamp":"` + ts + `"}}
`
	for _, u := range users {
		body += `{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":` + jsonQuote(u) + `}]}}
`
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write rollout: %v", err)
	}
	return path
}

func jsonQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

func TestCodexRolloutPromptMatches_SecondUserIsInject(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "sessions", "2026", "08", "13")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	ws := "/tmp/ws-agents"
	inject := "SeaTalk local-bot session open\nsession-id: s1\ncli: reply"
	path := writeRolloutUsers(t, day, "rollout-s-dddddddd-dddd-dddd-dddd-dddddddddddd.jsonl",
		"dddddddd-dddd-dddd-dddd-dddddddddddd", ws, []string{
			"# AGENTS.md instructions for /tmp/ws-agents\n\n<INSTRUCTIONS>\nbrief\n",
			inject,
		}, time.Now())
	if !codexRolloutPromptMatches(path, inject) {
		t.Fatal("want inject match on user[2] after AGENTS.md user[1]")
	}
	id, _, ok, err := scanActiveCodexTranscripts(root, ws, time.Now().Add(-time.Minute), inject)
	if err != nil || !ok || id != "dddddddd-dddd-dddd-dddd-dddddddddddd" {
		t.Fatalf("scan: ok=%v id=%q err=%v", ok, id, err)
	}
}

func TestCodexRolloutPromptMatches_StopsAfterThreeUsers(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "sessions", "2026", "08", "13")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	ws := "/tmp/ws-cap"
	path := writeRolloutUsers(t, day, "rollout-s-eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee.jsonl",
		"eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", ws, []string{
			"user-one",
			"user-two",
			"user-three",
			"SeaTalk local-bot session open",
		}, time.Now())
	if codexRolloutPromptMatches(path, "SeaTalk local-bot session open") {
		t.Fatal("inject at user[4] must not match (cap is 3)")
	}
	if !codexRolloutPromptMatches(path, "user-three") {
		t.Fatal("user[3] should still match")
	}
}

func TestFirstCodexRealUserPrompt_SkipsEnvironmentContext(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "sessions", "2026", "08", "12")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	ws := "/tmp/ws-test"
	path := writeRollout(t, day, "rollout-x-cccccccc-cccc-cccc-cccc-cccccccccccc.jsonl",
		"cccccccc-cccc-cccc-cccc-cccccccccccc", ws, "REAL_PROMPT", time.Now())
	got, ok := firstCodexRealUserPrompt(path)
	if !ok || got != "REAL_PROMPT" {
		t.Fatalf("got ok=%v text=%q", ok, got)
	}
}

func TestFindCodexTranscriptBySessionID_globAndSkip(t *testing.T) {
	root := t.TempDir()
	id := "01a018f9-a81b-76e1-84df-9c7ae9a054ef"
	day := filepath.Join(root, "sessions", "2026", "08", "19")
	if err := os.MkdirAll(day, 0755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(day, "rollout-2026-08-19T15-43-29-"+id+".jsonl")
	if err := os.WriteFile(want, []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	path, found, err := FindCodexTranscriptBySessionID(root, id)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if path != want {
		t.Fatalf("path=%q want %q", path, want)
	}

	_, found, err = FindCodexTranscriptBySessionID(root, "missing-session-id-00000000-0000-0000-0000-000000000000")
	if err != nil || found {
		t.Fatalf("missing: found=%v err=%v", found, err)
	}
	_, found, err = FindCodexTranscriptBySessionID(root, "")
	if err != nil || found {
		t.Fatalf("empty id: found=%v err=%v", found, err)
	}
}
