package agenttty

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCodexSubmitOverlap(t *testing.T) {
	sent := strings.Repeat("a", 100)
	if got := codexSubmitOverlap(sent, sent); got != 1 {
		t.Fatalf("identical = %v", got)
	}
	if got := codexSubmitOverlap(sent, sent+strings.Repeat("b", 15)); got < codexSubmitOverlapMin {
		t.Fatalf("15%% extra = %v, want >= %v", got, codexSubmitOverlapMin)
	}
	if got := codexSubmitOverlap(sent, sent+strings.Repeat("b", 40)); got >= codexSubmitOverlapMin {
		t.Fatalf("40%% extra = %v, want < %v", got, codexSubmitOverlapMin)
	}
	changed := []rune("abcdefghij")
	changed[0], changed[3], changed[7] = 'X', 'Y', 'Z'
	if got := codexSubmitOverlap("abcdefghij", string(changed)); got >= codexSubmitOverlapMin {
		t.Fatalf("3/10 edits = %v, want < %v", got, codexSubmitOverlapMin)
	}
	if got := codexSubmitOverlap("ok", "no"); got >= codexSubmitOverlapMin {
		t.Fatalf("ok/no = %v", got)
	}
	if got := codexSubmitOverlap("hello", "hallo"); got < codexSubmitOverlapMin {
		t.Fatalf("hello/hallo = %v, want >= %v", got, codexSubmitOverlapMin)
	}
	big := strings.Repeat("a", codexSubmitExactMaxBytes+64)
	if got := codexSubmitOverlap(big, big); got != 1 {
		t.Fatalf("large identical = %v", got)
	}
	mut := []rune(big)
	for i := range mut {
		if i%3 == 0 {
			mut[i] = 'b'
		}
	}
	if got := codexSubmitOverlap(big, string(mut)); got >= codexSubmitOverlapMin {
		t.Fatalf("large 1/3 edits = %v, want < %v", got, codexSubmitOverlapMin)
	}
}

func TestCodexPromptLanded_EnvironmentContextIsNotSubmit(t *testing.T) {
	home, ws := confirmFixture(t)
	env := "<environment_context>\n  <cwd>" + ws + "</cwd>\n</environment_context>"
	writeConfirmRollout(t, home, ws, "rollout-env.jsonl", []string{env})
	if codexPromptLanded(home, ws, env, map[string]int64{}) {
		t.Fatal("environment_context must not count as the injected prompt")
	}
}

func TestCodexPromptLanded_AfterBaseline(t *testing.T) {
	home, ws := confirmFixture(t)
	path := writeConfirmRollout(t, home, ws, "rollout-base.jsonl", []string{
		"SeaTalk local-bot session open\nsession-id: old",
	})
	base := CodexRolloutBaselines(home, ws)
	prompt := "SeaTalk local-bot session open\nsession-id: new"
	if codexPromptLanded(home, ws, prompt, base) {
		t.Fatal("history before baseline must not count")
	}
	appendConfirmUser(t, path, prompt)
	if !codexPromptLanded(home, ws, prompt, base) {
		t.Fatal("user line after baseline must count")
	}
}

func TestCodexPromptLanded_PartialLineIgnored(t *testing.T) {
	home, ws := confirmFixture(t)
	path := writeConfirmRollout(t, home, ws, "rollout-partial.jsonl", nil)
	base := CodexRolloutBaselines(home, ws)
	prompt := "SeaTalk local-bot session open"
	line := confirmUserLine(prompt)
	if err := os.WriteFile(path, append(mustRead(t, path), []byte(line)...), 0o644); err != nil {
		t.Fatal(err)
	}
	if codexPromptLanded(home, ws, prompt, base) {
		t.Fatal("line without newline must not count")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if !codexPromptLanded(home, ws, prompt, base) {
		t.Fatal("completed line must count")
	}
}

func TestCodexPromptLanded_LongLine(t *testing.T) {
	home, ws := confirmFixture(t)
	prompt := strings.Repeat("x", 600*1024) + "END"
	writeConfirmRollout(t, home, ws, "rollout-long.jsonl", []string{prompt})
	if !codexPromptLanded(home, ws, prompt, map[string]int64{}) {
		t.Fatal("user line over 512KB must still match")
	}
}

func TestCodexPromptLanded_OtherCwdIgnored(t *testing.T) {
	home, ws := confirmFixture(t)
	other := filepath.Join(home, "other-ws")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfirmRollout(t, home, other, "rollout-other.jsonl", []string{"SeaTalk local-bot session open"})
	if codexPromptLanded(home, ws, "SeaTalk local-bot session open", map[string]int64{}) {
		t.Fatal("other cwd must not count")
	}
}

func TestCodexSubmitConfirmLoop_OneEnterThenStop(t *testing.T) {
	polls := 0
	enters := 0
	runCodexSubmitConfirmLoop(context.Background(), codexSubmitLoop{
		interval: time.Millisecond,
		budget:   time.Minute,
		after: func(time.Duration) <-chan time.Time {
			ch := make(chan time.Time, 1)
			ch <- time.Now()
			return ch
		},
		landed: func() bool {
			polls++
			return polls >= 2
		},
		enter: func() error {
			enters++
			return nil
		},
	})
	if enters != 1 || polls != 2 {
		t.Fatalf("enters=%d polls=%d, want 1 and 2", enters, polls)
	}
}

func TestCodexSubmitConfirmLoop_AlreadyLandedNoEnter(t *testing.T) {
	enters := 0
	runCodexSubmitConfirmLoop(context.Background(), codexSubmitLoop{
		interval: time.Millisecond,
		budget:   time.Minute,
		after: func(time.Duration) <-chan time.Time {
			ch := make(chan time.Time, 1)
			ch <- time.Now()
			return ch
		},
		landed: func() bool { return true },
		enter: func() error {
			enters++
			return nil
		},
	})
	if enters != 0 {
		t.Fatalf("enters=%d, want 0", enters)
	}
}

func confirmFixture(t *testing.T) (home, workspace string) {
	t.Helper()
	home = t.TempDir()
	workspace = filepath.Join(home, "ws")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	return home, workspace
}

func writeConfirmRollout(t *testing.T, home, cwd, name string, users []string) string {
	t.Helper()
	day := filepath.Join(home, "sessions", "2026", "08", "12")
	if err := os.MkdirAll(day, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(day, name)
	var b strings.Builder
	b.WriteString(`{"timestamp":"2026-08-12T00:00:00Z","type":"session_meta","payload":{"id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","session_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","cwd":`)
	cwdJSON, _ := json.Marshal(cwd)
	b.Write(cwdJSON)
	b.WriteString(`,"timestamp":"2026-08-12T00:00:00Z"}}` + "\n")
	for _, u := range users {
		b.WriteString(confirmUserLine(u))
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func appendConfirmUser(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write([]byte(confirmUserLine(text) + "\n")); err != nil {
		t.Fatal(err)
	}
}

func confirmUserLine(text string) string {
	body, _ := json.Marshal(text)
	return `{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":` + string(body) + `}]}}`
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
