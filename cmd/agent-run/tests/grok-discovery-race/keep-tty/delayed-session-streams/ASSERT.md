---
label: e2e
---

## Expected

- `events.jsonl` contains `think` with text `Resolve session id...`.
- `events.jsonl` contains assistant `message` with `DELAYED_SESSION_MARKER` from delayed `updates.jsonl`.
- No `error` containing `context canceled` within 10s after the resolve think event.
- Stderr contains discovered grok session id and updates path once session dir appears.

## Side Effects

- Delayed grok session dir created under `GROK_HOME` after ~8s from grok-tty session registration.

## Errors

- Must not emit `Cannot resolve session id: context canceled` in the early discovery window (~1–2s bug repro).

## Exit Code

0 (run completes after streamed marker satisfies keep-tty wait)

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
	if len(resp.EventsParsed) == 0 {
		t.Fatalf("events.jsonl empty; stderr:\n%s", resp.Stderr)
	}
	if !eventsHaveThinkText(resp.EventsParsed, resolveThinkText) {
		t.Fatalf("events.jsonl missing think %q; events=%v", resolveThinkText, resp.EventsParsed)
	}
	if resp.EarlyContextCancel {
		t.Fatalf("early context canceled error within %s after think (bug repro); gap=%s events=%v stderr:\n%s",
			earlyCancelGuardWindow, resp.ThinkToErrorGap, resp.EventsParsed, resp.Stderr)
	}
	if !resp.HasDelayedMarker {
		t.Fatalf("expected assistant message containing %q in events.jsonl; events=%v stdout:\n%s stderr:\n%s",
			delayedSessionMarker, resp.EventsParsed, resp.Stdout, resp.Stderr)
	}
	wantSession := "grok-tty: grok session " + delayedSessionGrokUUID
	if !strings.Contains(resp.Stderr, wantSession) {
		t.Fatalf("stderr missing %q; stderr:\n%s", wantSession, resp.Stderr)
	}
	if req.GrokUpdatesPath != "" && !strings.Contains(resp.Stderr, req.GrokUpdatesPath) {
		t.Fatalf("stderr missing updates path %q; stderr:\n%s", req.GrokUpdatesPath, resp.Stderr)
	}
}
```