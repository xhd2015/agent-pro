---
label: e2e, slow
explanation: empty-home bind may wait full 90s discovery window after fix
---

## Expected

- `events.jsonl` contains `think` with text `Resolve session id...`.
- `events.jsonl` contains `error` whose text starts with `Cannot resolve session id:`.
- Think→error gap is **greater than 3s** (discovery kept polling past chrome false-complete window).
- If error mentions `context canceled`, it must not appear at ~1.2s (early cancel bug).

## Side Effects

- No grok session dir created under empty `GROK_HOME`.
- No assistant scrollback fallback message in `events.jsonl`.

## Errors

- Bind failure is expected; early `context canceled` at ~1s is the bug under test.

## Exit Code

Non-zero or finished with error status is acceptable; timing assertion is primary.

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
		t.Fatalf("events.jsonl empty after %s; stderr:\n%s", resp.Elapsed, resp.Stderr)
	}
	if !eventsHaveThinkText(resp.EventsParsed, resolveThinkText) {
		t.Fatalf("events.jsonl missing think %q; events=%v", resolveThinkText, resp.EventsParsed)
	}
	if !eventsHaveErrorPrefix(resp.EventsParsed, resolveErrorPrefix) {
		t.Fatalf("events.jsonl missing error prefix %q; events=%v stderr:\n%s",
			resolveErrorPrefix, resp.EventsParsed, resp.Stderr)
	}
	if resp.ThinkToErrorGap <= discoveryMinWindow {
		t.Fatalf("think→error gap %s must exceed discovery window %s (early chrome false-complete bug); events=%v stderr:\n%s",
			resp.ThinkToErrorGap, discoveryMinWindow, resp.EventsParsed, resp.Stderr)
	}
	if resp.HasContextCancel && resp.ThinkToErrorGap < discoveryMinWindow {
		t.Fatalf("context canceled appeared too early (%s); want >%s; stderr:\n%s",
			resp.ThinkToErrorGap, discoveryMinWindow, resp.Stderr)
	}
	for _, ev := range resp.EventsParsed {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			text, _ := ev["text"].(string)
			if strings.TrimSpace(text) != "" {
				t.Fatalf("unexpected assistant scrollback fallback in events.jsonl: %q", text)
			}
		}
	}
}
```