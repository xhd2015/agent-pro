## Expected

- Combined PTY/host capture contains no line starting with `session-`.
- Host must not print session id on stdout or stderr during attach or Ctrl-] detach.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	assertNoHostSessionID(t, combined)
}
```