## Expected

- Paris after open.
- Resume while live fails with active/send gate, **not** already-in-use.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris")
	}
	if resp.AlreadyInUse {
		t.Fatalf("must not report already in use while live; stderr=%s", resp.Resume.Stderr)
	}
	c := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	if strings.Contains(c, "already in use") {
		t.Fatalf("already in use (wrong error):\n%s", c)
	}
	if resp.Resume.ExitCode == 0 {
		t.Fatalf("resume while live should fail; got exit 0\n%s", c)
	}
	if !resp.ResumeDenied &&
		!strings.Contains(c, "still active") &&
		!strings.Contains(c, "use send") &&
		!strings.Contains(c, "not exited") &&
		!strings.Contains(c, "cannot resume") {
		t.Fatalf("want still-active/use-send style error; got:\n%s", c)
	}
}
```
