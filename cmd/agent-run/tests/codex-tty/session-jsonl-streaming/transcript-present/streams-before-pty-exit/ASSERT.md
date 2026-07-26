---
label: e2e
---

## Expected

- Exit code 0.
- `StreamProbeSeen` is true.
- `StreamProbeBeforeExit` is true: stdout received `JSONL_STREAM_BEFORE_PTY_EXIT`
  before fake Codex exited.

## Side Effects

- The marker is persisted in the codex-tty session events.

## Errors

- No error is expected.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if !resp.StreamProbeSeen {
		t.Fatalf("expected streamed marker %q on stdout; stdout:\n%s\nstderr:\n%s",
			codexPreExitStreamText, resp.Stdout, resp.Stderr)
	}
	if !resp.StreamProbeBeforeExit {
		t.Fatalf("expected marker while PTY was still running; stdout:\n%s", resp.Stdout)
	}
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, codexPreExitStreamText) {
		t.Fatalf("expected pre-exit streamed marker in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}
}
```
