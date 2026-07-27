## Expected

- Exit code 0.
- Stdout contains `JSONL_CURRENT_CWD_SHOULD_STREAM`.
- Stdout does not contain `JSONL_STALE_CWD_SHOULD_NOT_STREAM`.

## Side Effects

- Persisted events include only the current assistant message from the selected transcript.

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
		t.Fatalf("expected current cwd marker %q on stdout; stdout:\n%s\nstderr:\n%s",
			codexCurrentCWDText, resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stdout, codexStaleCWDText) {
		t.Fatalf("stale same-cwd transcript should not be streamed; stdout:\n%s", resp.Stdout)
	}
	_, lines := findCodexTTYEventsJSONL(t, req.Home)
	joined := strings.Join(lines, "\n")
	if !eventsContainSubstring(t, lines, codexCurrentCWDText) {
		t.Fatalf("expected current cwd marker in events.jsonl:\n%s", joined)
	}
	if strings.Contains(joined, codexStaleCWDText) {
		t.Fatalf("stale same-cwd marker should not be persisted:\n%s", joined)
	}
}
```
