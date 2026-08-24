## Expected

- Row contains session id, `w=3 t=1`, TITLE header (no title on disk → `-`), workspace.
- No `SENDABLE` column.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	out := resp.Stdout
	if strings.Contains(out, "SENDABLE") {
		t.Fatalf("SENDABLE column must be gone:\n%s", out)
	}
	for _, want := range []string{"TITLE", fixtureListLiveSID, "w=3 t=1", "/tmp/proj", "1 sessions"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout missing %q:\n%s", want, out)
		}
	}
}
```
