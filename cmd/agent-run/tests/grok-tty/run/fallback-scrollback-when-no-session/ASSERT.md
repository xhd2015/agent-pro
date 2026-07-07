## Expected

- Exit code 0.
- Stderr contains a warning that grok session discovery failed (mentions grok session /
  updates / discovery — not only the `grok-tty: session-N` registry line).
- `events.jsonl` contains `error` event with prefix `Cannot resolve session id:`.
- No scrollback-captured assistant text `hi` in stdout or `events.jsonl`.

## Exit Code

0

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

const resolveErrorPrefix = "Cannot resolve session id:"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	stderrLower := strings.ToLower(resp.Stderr)
	hasWarning := strings.Contains(stderrLower, "grok session") ||
		strings.Contains(stderrLower, "updates.jsonl") ||
		strings.Contains(stderrLower, "discovery")
	if !hasWarning {
		t.Fatalf("expected stderr warning when grok session dir missing; stderr:\n%s", resp.Stderr)
	}

	_, lines := findGrokTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, resolveErrorPrefix) {
		t.Fatalf("events.jsonl missing error prefix %q:\n%s", resolveErrorPrefix, strings.Join(lines, "\n"))
	}
	for _, line := range lines {
		var ev map[string]any
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		text, _ := ev["text"].(string)
		trimmed := strings.TrimSpace(text)
		if trimmed == "hi" || strings.Contains(text, "Response: hi") {
			t.Fatalf("scrollback fallback text should not appear in events.jsonl:\n%s", strings.Join(lines, "\n"))
		}
	}
	if strings.Contains(resp.Stdout, "hi") {
		t.Fatalf("scrollback fallback hi should not appear on stdout:\n%s", resp.Stdout)
	}
}
```