## Expected

- Live PID but no iTerm match → warning + resume.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.ListITermCalls != 1 {
		t.Fatalf("ListITermCalls = %d, want 1", resp.ListITermCalls)
	}
	if !strings.Contains(resp.Stderr, "warning:") || !strings.Contains(resp.Stderr, "no matching iTerm") {
		t.Fatalf("stderr missing live-no-iterm warning:\n%s", resp.Stderr)
	}
	assertResumeOpened(t, req, resp)
	if !strings.Contains(resp.Stdout, "opened: new window; resuming "+req.SessionID) {
		t.Fatalf("stdout = %q", resp.Stdout)
	}
}
```
