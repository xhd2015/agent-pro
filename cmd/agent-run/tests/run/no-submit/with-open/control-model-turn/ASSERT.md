---
label: e2e, real-grok, slow
explanation: Control oracle for Option C; real grok must produce mock HTTP or session user_message.
---

## Expected

- Exit code 0.
- Stderr has a `grok-tty: <terminal-id>` line.
- **Turn evidence (hard)**: after settle, **at least one** of:
  - mock HTTP chat log contains a turn for the prompt, or
  - GROK_HOME session files contain a `user_message` (or equivalent) for the prompt.
- Proves `no-model-turn` is not a false green from a dead mock/harness.

## Side Effects

- Keep-alive TTY + mock server may remain until cleanup.
- Provider session may bind successfully (hard discovery with config home).

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("control leaf exit code = %d, stderr:\n%s\nstdout:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	if _, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty"); !ok {
		t.Fatalf("missing post-attach grok-tty terminal session id on stderr:\n%s", resp.Stderr)
	}

	if !resp.HasMockChatHTTP && !resp.HasUserMessageForPrompt {
		t.Fatalf("control leaf expected turn evidence (mock HTTP and/or session user_message) for prompt %q\n"+
			"HasMockChatHTTP=%v HasUserMessageForPrompt=%v sessionsScanned=%d log-http lines=%d\n"+
			"log-http:\n%s\ngrokHome=%s\nstderr:\n%s",
			req.Prompt,
			resp.HasMockChatHTTP, resp.HasUserMessageForPrompt, resp.GrokSessionsScanned, len(resp.LogHTTPLines),
			trimControlLog(resp.LogHTTPContent, 4000),
			req.GrokHome, resp.Stderr)
	}
}

func trimControlLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(empty log-http file)"
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n…(truncated)…"
}
```
