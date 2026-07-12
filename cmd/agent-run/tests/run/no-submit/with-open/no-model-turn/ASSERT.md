---
label: real-grok, slow
explanation: Option C real grok TUI under llm-mock-run-grok; no-turn oracle needs live grok + mock HTTP settle.
---

## Expected

- Exit code **0** (open lifecycle completes; soft unbound allowed when no provider
  session exists because the draft was never submitted).
- Stderr has exactly one `grok-tty: <terminal-id>` line (post-attach).
- Must **not** fail solely as `unrecognized flag` / `unknown flag` for `--no-submit`.
- Must **not** hard-fail solely with `session id not resolved` / `grok session id
  not resolved` as the only reason for non-zero exit under `--no-submit` (bind
  policy A: soft unbound). When exit is non-zero for that reason alone, treat as
  product gap (RED until soft unbound lands with NoSubmit).
- **Primary oracle (hard)**: after settle, **no** mock HTTP chat turn for the
  draft **and** **no** session `user_message` (updates/events) containing the
  draft. Either form of turn evidence is a failure.
- Soft: draft may appear in PTY input buffer / snapshot (optional; not required).

## Side Effects

- Keep-alive TTY session may remain (registry/listen_addr); cleanup kills serve
  PIDs after the test.
- `LLM_MOCK_RUN_FLAGS --log-http` file may be empty or absent of chat lines.

## Errors

- None expected on the happy path after product fix (argv + inject + soft unbound).

## Exit Code

0

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}

	combinedCLI := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(combinedCLI, "unrecognized flag") || strings.Contains(combinedCLI, "unknown flag") {
		t.Fatalf("want implemented --no-submit path, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}

	// Primary Option C oracle first (even when exit ≠ 0): argv auto-submit shows as
	// mock HTTP / session user_message for the draft. Prefer this RED reason when present.
	var turnProblems []string
	if resp.HasMockChatHTTP {
		turnProblems = append(turnProblems, fmt.Sprintf(
			"mock HTTP shows chat for draft %q (%d log-http lines):\n%s",
			req.Prompt, len(resp.LogHTTPLines), trimForFail(resp.LogHTTPContent, 4000)))
	}
	if resp.HasUserMessageForPrompt {
		turnProblems = append(turnProblems, fmt.Sprintf(
			"session user_message for draft %q (scanned %d session file(s) under %s)",
			req.Prompt, resp.GrokSessionsScanned, req.GrokHome))
	}
	if len(turnProblems) > 0 {
		t.Fatalf("--no-submit must not start a model turn:\n- %s\nstderr:\n%s",
			strings.Join(turnProblems, "\n- "), resp.Stderr)
	}

	// Soft-unbound contract: non-zero solely for unresolved bind under --no-submit
	// is still RED (product must soft-unbound when draft was never submitted).
	if resp.ExitCode != 0 {
		if strings.Contains(combinedCLI, "session id not resolved") ||
			strings.Contains(combinedCLI, "grok session id not resolved") {
			t.Fatalf("--no-submit must soft-unbound (exit 0) when provider session is missing; got exit %d:\n"+
				"(turn oracle clean: HasMockChatHTTP=%v HasUserMessage=%v sessionsScanned=%d log-http lines=%d)\nstderr:\n%s\nstdout:\n%s",
				resp.ExitCode, resp.HasMockChatHTTP, resp.HasUserMessageForPrompt, resp.GrokSessionsScanned, len(resp.LogHTTPLines),
				resp.Stderr, resp.Stdout)
		}
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty")
	if !ok {
		t.Fatalf("missing post-attach grok-tty terminal session id on stderr:\n%s", resp.Stderr)
	}
	if n := countPrefixedSessionIDLines(resp.Stderr, "grok-tty"); n != 1 {
		t.Fatalf("want exactly 1 grok-tty session id line, got %d (id=%q)\nstderr:\n%s", n, id, resp.Stderr)
	}
}

func trimForFail(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n…(truncated)…"
}
```
