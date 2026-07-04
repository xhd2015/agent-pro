---
label: real-grok, slow
explanation: Requires real grok CLI on PATH; reproduces interactive no-config hello hang.
---

## Expected

- Exit code 0 within 30 seconds (grok must not hang waiting for mock).
- Combined stdout/stderr contains assistant-visible text beyond `GROK_HOME=` alone.
- Newest grok session `events.jsonl` contains `first_token` (model produced output).

## Side Effects

- `$GROK_HOME/sessions/<encoded-workdir>/<uuid>/events.jsonl` written by grok.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	combined := resp.Stdout + resp.Stderr
	trimmed := strings.TrimSpace(strings.ReplaceAll(combined, "GROK_HOME="+resp.GrokHomeUsed, ""))
	if trimmed == "" {
		t.Fatalf("expected assistant-visible output for prompt hello, got only GROK_HOME announcement:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}

	if resp.GrokEventsPath == "" {
		t.Fatalf("expected grok session events.jsonl; GrokHomeUsed=%q stdout:\n%s stderr:\n%s",
			resp.GrokHomeUsed, resp.Stdout, resp.Stderr)
	}
	if !grokEventsHasEventType(resp.GrokEventsLines, "first_token") {
		t.Fatalf("events.jsonl missing first_token (grok stuck waiting for mock LLM?):\n%s",
			strings.Join(resp.GrokEventsLines, "\n"))
	}
	if !grokEventsHasTurnStarted(resp.GrokEventsLines, "mock-model") {
		t.Fatalf("events.jsonl missing turn_started with model_id mock-model:\n%s",
			strings.Join(resp.GrokEventsLines, "\n"))
	}
}
```