---
label: e2e, real-grok, slow
explanation: Requires real grok CLI on PATH; headless LLM round-trip ~tens of seconds.
---

## Expected

- Exit code 0.
- Combined stdout/stderr contains `"Paris"` (mocked LLM response).
- Newest grok session `events.jsonl` exists under encoded workdir.
- `events.jsonl` contains `turn_started` event with `model_id: "mock-model"`.

## Side Effects

- `$GROK_HOME/sessions/<encoded-workdir>/<uuid>/events.jsonl` written by grok.

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
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "Paris")

	if resp.GrokEventsPath == "" {
		t.Fatalf("expected grok session events.jsonl; GrokHomeUsed=%q stdout:\n%s stderr:\n%s",
			resp.GrokHomeUsed, resp.Stdout, resp.Stderr)
	}
	if len(resp.GrokEventsLines) == 0 {
		t.Fatalf("expected non-empty events.jsonl at %s", resp.GrokEventsPath)
	}
	if !grokEventsHasTurnStarted(resp.GrokEventsLines, "mock-model") {
		t.Fatalf("events.jsonl missing turn_started with model_id mock-model:\n%s",
			strings.Join(resp.GrokEventsLines, "\n"))
	}
}
```
