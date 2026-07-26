# Scenario

**Feature**: agent-pro codex sessions list, brief, and log from rollout JSONL

```
# harness builds synthetic Codex home with rollout transcripts
test harness -> sessions package -> rollout files under CodexHome/sessions/

# list discovers sessions; brief summarizes; log prints compact trace lines
sessions package -> print.FormatTraceLine pipeline -> formatted terminal output
```

## Preconditions

- Package `agent/codex/sessions` exposes List, Find, Brief, PrintLog, and
  table/JSON formatters.
- Tests never read the real user `~/.codex` directory.

## Steps

1. Root Setup allocates `req.CodexHome` as `{temp}/.codex`.
2. Root Setup sets `req.Now` to a fixed UTC instant for deterministic relative times.
3. Leaf Setup writes rollout JSONL fixtures under `CodexHome/sessions/`.
4. Run dispatches to List, Info, Brief, or PrintLog based on `req.Operation`.

## Context

- Rollout path pattern:
  `{CodexHome}/sessions/YYYY/MM/DD/rollout-<ts>-<uuid>.jsonl`
- Codex trace adapter must be registered for log formatting.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/xhd2015/agent-pro/agent_trace/codex"
)

const fixedNow = "2026-07-03T15:00:00.000Z"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CodexHome = filepath.Join(t.TempDir(), ".codex")
	now, err := time.Parse(time.RFC3339, fixedNow)
	if err != nil {
		t.Fatalf("parse fixed now: %v", err)
	}
	req.Now = now.UTC()
	return nil
}

func writeRolloutSession(t *testing.T, codexHome, id, timestamp, cwd string, extraLines ...string) string {
	t.Helper()
	ts, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp)
	if err != nil {
		ts, err = time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Fatalf("parse timestamp %q: %v", timestamp, err)
		}
	}
	meta := fmt.Sprintf(
		`{"type":"session_meta","payload":{"id":%q,"timestamp":%q,"cwd":%q}}`,
		id, timestamp, cwd,
	)
	lines := append([]string{meta}, extraLines...)
	dir := filepath.Join(
		codexHome, "sessions",
		ts.UTC().Format("2006"), ts.UTC().Format("01"), ts.UTC().Format("02"),
	)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir rollout dir: %v", err)
	}
	filename := fmt.Sprintf("rollout-%s-%s.jsonl", ts.UTC().Format("2006-01-02T15-04-05"), id)
	path := filepath.Join(dir, filename)
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write rollout %s: %v", path, err)
	}
	return path
}

func agentMessageLine(message string) string {
	return fmt.Sprintf(
		`{"type":"event_msg","payload":{"type":"agent_message","message":%q,"phase":"commentary"}}`,
		message,
	)
}

func userMessageLine(message string) string {
	return fmt.Sprintf(
		`{"type":"event_msg","payload":{"type":"user_message","message":%q}}`,
		message,
	)
}

func tokenCountLine(inputTokens, outputTokens int) string {
	return fmt.Sprintf(
		`{"type":"event_msg","payload":{"type":"token_count","input_tokens":%d,"output_tokens":%d}}`,
		inputTokens, outputTokens,
	)
}

func execCommandLines(cmd, output, callID string) []string {
	return []string{
		fmt.Sprintf(
			`{"type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":%q,"call_id":%q}}`,
			fmt.Sprintf(`{"cmd":%q}`, cmd), callID,
		),
		fmt.Sprintf(
			`{"type":"response_item","payload":{"type":"function_call_output","call_id":%q,"output":%q}}`,
			callID, output,
		),
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("operation failed: %v", resp.Err)
	}
}

func assertError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error but got nil")
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}

```