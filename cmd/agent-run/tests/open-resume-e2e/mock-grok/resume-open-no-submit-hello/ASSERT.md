## Expected

- Full open/exit path succeeds.
- Resume with `--open --no-submit` does not error about missing `--open`.
- Not already-in-use; resume exit 0 preferred.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatalf("flow error: %v", resp.Err)
	}
	if !resp.HasParis || !resp.ExitedTrue {
		t.Fatalf("need Paris+exited; paris=%v exited=%v status=\n%s",
			resp.HasParis, resp.ExitedTrue, resp.StatusAfterExit.Stdout)
	}
	c := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	if strings.Contains(c, "no-submit requires --open") || strings.Contains(c, "--no-submit requires") {
		t.Fatalf("unexpected no-submit gate failure:\n%s", c)
	}
	if resp.AlreadyInUse || strings.Contains(c, "already in use") {
		t.Fatalf("already in use:\n%s", c)
	}
	if resp.Resume.ExitCode != 0 {
		t.Fatalf("resume exit=%d:\n%s", resp.Resume.ExitCode, c)
	}
}
```
