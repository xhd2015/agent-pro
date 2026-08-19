---
label: e2e
---

## Expected

- Exit code 0.
- Stderr contains a warning that grok session discovery failed (mentions grok session /
  updates / discovery — not only the `grok-tty: session-N` registry line).
- `events.jsonl` contains `error` event with prefix `Cannot resolve session id:`.
- Scrollback fallback surfaces the fake-TUI response (`hi` / `Response: hi`) on
  stdout and/or as an assistant event (same as codex/commandcode when transcript
  tail never streams).

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

const resolveErrorPrefix = "Cannot resolve session id:"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	stderrLower := strings.ToLower(resp.Stderr)
	hasWarning := strings.Contains(stderrLower, "grok session") ||
		strings.Contains(stderrLower, "updates.jsonl") ||
		strings.Contains(stderrLower, "discovery") ||
		strings.Contains(stderrLower, "scrollback")
	if !hasWarning {
		t.Fatalf("expected stderr warning when grok session dir missing; stderr:\n%s", resp.Stderr)
	}

	_, lines := findGrokTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, resolveErrorPrefix) {
		t.Fatalf("events.jsonl missing error prefix %q:\n%s", resolveErrorPrefix, strings.Join(lines, "\n"))
	}
	hasHi := strings.Contains(resp.Stdout, "hi")
	if !hasHi {
		for _, line := range lines {
			var ev map[string]any
			if json.Unmarshal([]byte(line), &ev) != nil {
				continue
			}
			text, _ := ev["text"].(string)
			if strings.TrimSpace(text) == "hi" || strings.Contains(text, "Response: hi") {
				hasHi = true
				break
			}
		}
	}
	if !hasHi {
		t.Fatalf("expected scrollback fallback hi on stdout or events; stdout:\n%s\nevents:\n%s",
			resp.Stdout, strings.Join(lines, "\n"))
	}
}
```
