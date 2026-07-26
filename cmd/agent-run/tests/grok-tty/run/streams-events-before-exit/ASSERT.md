---
label: e2e
---

## Expected

- Exit code 0.
- `StreamProbeSeen` is true: stdout contains `STREAM_PROBE_LS_DONE` **before** the fake
  TUI exits (while the PTY session is still running).
- Stdout shows streamed tool output (`agent`) or formatted tool line, not only the
  final scrollback blob at process exit.

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
	if !resp.StreamProbeSeen {
		t.Fatalf("expected streamed marker %q on stdout before timeout; stdout:\n%s\nstderr:\n%s",
			streamProbeMarker, resp.Stdout, resp.Stderr)
	}
	if !resp.StreamProbeBeforeExit {
		t.Fatalf("expected marker while PTY still running (before fake TUI exit); stdout:\n%s", resp.Stdout)
	}
	combined := strings.ToLower(resp.Stdout)
	if !strings.Contains(combined, "stream_probe_ls_done") && !strings.Contains(combined, "agent") {
		t.Fatalf("expected streamed assistant/tool content on stdout:\n%s", resp.Stdout)
	}
}
```