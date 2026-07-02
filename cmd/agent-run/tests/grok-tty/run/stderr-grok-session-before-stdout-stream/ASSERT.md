## Expected

- Exit code 0.
- `StreamProbeSeen` is true: stdout contains `STREAM_ORDER_MARKER` before fake TUI exits.
- Stderr contains `grok-tty: grok session <uuid>` and `grok-tty: grok updates <path>`.
- Stderr grok session lines appear **before** the first stdout stream marker (ordering).

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
		t.Fatalf("expected streamed marker %q on stdout before timeout; stdout:\n%s\nstderr:\n%s",
			orderProbeMarker, resp.Stdout, resp.Stderr)
	}
	assertStderrGrokSession(t, resp.Stderr, orderProbeGrokUUID, req.GrokUpdatesPath)
	assertStderrGrokSessionBeforeStdout(t, resp.Stderr, resp.Stdout, orderProbeMarker, orderProbeGrokUUID)
}
```