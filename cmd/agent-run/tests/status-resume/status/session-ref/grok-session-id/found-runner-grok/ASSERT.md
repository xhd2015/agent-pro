## Expected

- Exit code 0.
- Stdout identifies `grok/test-gsid-grok-s1` (or bare session id) and the UUID.
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
	assertContainsAny(t, resp.Stdout,
		req.Runner+"/"+req.SessionID,
		req.SessionID,
	)
	assertContains(t, resp.Stdout, req.RunnerSessionID)
	assertTrailingNewline(t, resp.Stdout, "status stdout")
}
```
