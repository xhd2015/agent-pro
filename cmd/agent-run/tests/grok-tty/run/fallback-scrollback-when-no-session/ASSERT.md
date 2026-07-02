## Expected

- Exit code 0.
- Stderr contains a warning that grok session discovery failed (mentions grok session /
  updates / scrollback fallback — not only the `grok-tty: session-N` registry line).
- Stdout or `events.jsonl` still contains scrollback-captured assistant text `hi`.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	stderrLower := strings.ToLower(resp.Stderr)
	hasWarning := strings.Contains(stderrLower, "grok session") ||
		strings.Contains(stderrLower, "updates.jsonl") ||
		strings.Contains(stderrLower, "scrollback")
	if !hasWarning {
		t.Fatalf("expected stderr warning when grok session dir missing; stderr:\n%s", resp.Stderr)
	}

	if !strings.Contains(resp.Stdout, "hi") {
		t.Fatalf("expected scrollback fallback hi on stdout:\n%s", resp.Stdout)
	}
	_, lines := findGrokTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, "hi") {
		t.Fatalf("expected scrollback fallback hi in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}

	assert.Output(t, resp.Stdout, `
<contains>
hi
</contains>`)
}
```