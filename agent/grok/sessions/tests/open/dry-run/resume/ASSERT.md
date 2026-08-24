## Expected

- Dry-run resume prints plan; OpenInNewWindow not called.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertNoSideEffects(t, resp)
	out := resp.Stdout
	for _, want := range []string{"Would open new iTerm2 window", req.SessionID, req.ProjectDir, "--resume"} {
		if !strings.Contains(out, want) {
			t.Fatalf("dry-run stdout missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "--fork-session") {
		t.Fatalf("dry-run must not fork:\n%s", out)
	}
}
```
