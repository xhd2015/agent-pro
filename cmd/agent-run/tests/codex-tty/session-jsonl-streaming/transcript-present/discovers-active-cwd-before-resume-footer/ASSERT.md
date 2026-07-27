## Expected

- Exit code 0.
- `StreamProbeSeen` is true.
- `StreamProbeBeforeExit` is true: stdout received
  `JSONL_ACTIVE_CWD_BEFORE_RESUME_FOOTER` while fake Codex was still running.

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
		t.Fatalf("expected active cwd streamed marker %q on stdout before resume footer; stdout:\n%s\nstderr:\n%s",
			codexActiveCWDText, resp.Stdout, resp.Stderr)
	}
	if !resp.StreamProbeBeforeExit {
		t.Fatalf("expected active cwd marker while PTY was still running; stdout:\n%s", resp.Stdout)
	}
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	if !eventsContainSubstring(t, lines, codexActiveCWDText) {
		t.Fatalf("expected active cwd streamed marker in events.jsonl:\n%s", strings.Join(lines, "\n"))
	}
}
```
