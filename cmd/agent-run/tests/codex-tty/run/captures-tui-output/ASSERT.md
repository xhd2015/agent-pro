## Expected

- Exit code 0.
- Stdout contains captured assistant text `hi` (from `Response: hi`).
- `AGENT_RUN_HOME/sessions/codex-tty/<id>/events.jsonl` exists and contains `hi`
  in an assistant/message event.

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
	if !strings.Contains(resp.Stdout, "hi") {
		t.Fatalf("stdout missing captured response hi:\n%s", resp.Stdout)
	}
	path, lines := findCodexTTYEventsJSONL(t, req.Home)
	if path == "" {
		t.Fatal("events.jsonl not found under sessions/codex-tty/")
	}
	if !eventsContainSubstring(t, lines, "hi") {
		t.Fatalf("events.jsonl missing hi:\n%s", strings.Join(lines, "\n"))
	}
}
```