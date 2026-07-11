## Expected

- Exit code 0.
- Stdout identifies the session as `grok-tty/test-ref-s1` (or equivalent).
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
	assertContains(t, resp.Stdout, req.Runner+"/"+req.SessionID)
	assertTrailingNewline(t, resp.Stdout, "status stdout")
}
```
