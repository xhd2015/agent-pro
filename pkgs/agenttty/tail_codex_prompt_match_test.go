package agenttty

import (
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
