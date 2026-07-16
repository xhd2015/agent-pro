## Expected

- Exit code 0.
- Stdout identifies the agent-run session (`grok-tty/test-gsid-s1` or bare id
  plus runner context).
- Stdout includes the provider UUID (bound runner session id).
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	// Prefer compound display form used by multi-layer status; bare id also ok.
	assertContainsAny(t, resp.Stdout,
		req.Runner+"/"+req.SessionID,
		req.SessionID,
	)
	assertContains(t, resp.Stdout, req.RunnerSessionID)
	assertTrailingNewline(t, resp.Stdout, "status stdout")
}
```
