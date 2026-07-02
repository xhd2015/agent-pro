## Expected

- Exit code 0.
- Grok session dir did not exist at run start; created 5s after prompt inject.
- `StreamProbeSeen` is true: stdout contains `DELAYED_SESSION_MARKER` while PTY still running.
- Stderr contains discovered grok session id and updates path (proves polling continued until session appeared).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if !resp.StreamProbeSeen {
		t.Fatalf("expected delayed session marker %q on stdout; stdout:\n%s\nstderr:\n%s",
			delayedSessionMarker, resp.Stdout, resp.Stderr)
	}
	if !resp.StreamProbeBeforeExit {
		t.Fatalf("expected marker while PTY still running (before fake TUI exit); stdout:\n%s", resp.Stdout)
	}
	assertStderrGrokSession(t, resp.Stderr, delayedSessionGrokUUID, req.GrokUpdatesPath)
}
```