## Expected

- Exit code 0 within 5 seconds (no `ExecTimeout` context deadline exceeded).
- Orchestrator returns promptly after grok exits even when session dir exists without `events.jsonl`.
- stderr must not report mirror-not-ready after a multi-second stall (optional warning OK if fast).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp.ExitCode == -1 {
		t.Fatalf("orchestrator hung after grok exit (ExecTimeout %s exceeded; likely waitAndMirrorSessions 60s poll when session lacks events.jsonl)\nstdout:\n%s\nstderr:\n%s",
			req.ExecTimeout, resp.Stdout, resp.Stderr)
	}
	if resp.Err != nil {
		if strings.Contains(resp.Err.Error(), "timeout") || strings.Contains(resp.Err.Error(), "deadline") {
			t.Fatalf("orchestrator hung after grok exit: %v\nstdout:\n%s\nstderr:\n%s",
				resp.Err, resp.Stdout, resp.Stderr)
		}
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
}
```