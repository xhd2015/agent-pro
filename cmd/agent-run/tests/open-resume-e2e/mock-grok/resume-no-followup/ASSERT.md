## Expected

- Paris + exited true after /exit.
- Resume without followup does **not** say prompt is required.
- Resume not already-in-use; exit 0 preferred.

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
		t.Fatalf("flow error: %v\n%s", resp.Err, resp.Open.Stderr)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris; snap=%s", resp.ParisSnapshot)
	}
	if !resp.ExitedTrue {
		t.Fatalf("want exited true; status=\n%s", resp.StatusAfterExit.Stdout)
	}
	c := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	if strings.Contains(c, "prompt is required") || strings.Contains(c, "prompt required") {
		t.Fatalf("resume no-followup must not require prompt:\n%s", c)
	}
	if resp.AlreadyInUse || strings.Contains(c, "already in use") {
		t.Fatalf("already in use:\n%s", c)
	}
	if resp.Resume.ExitCode != 0 {
		t.Fatalf("resume exit=%d want 0:\n%s", resp.Resume.ExitCode, c)
	}
}
```
