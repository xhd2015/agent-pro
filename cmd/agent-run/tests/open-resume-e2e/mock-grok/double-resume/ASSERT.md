## Expected

- First cycle: Paris, exited, resume ok.
- Second resume (Resume2) exit 0; not already-in-use.

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
		t.Fatalf("flow: %v", resp.Err)
	}
	if !resp.HasParis || !resp.ExitedTrue {
		t.Fatalf("first cycle incomplete paris=%v exited=%v", resp.HasParis, resp.ExitedTrue)
	}
	if resp.AlreadyInUse {
		t.Fatalf("already in use on double resume")
	}
	if resp.Resume.ExitCode != 0 {
		t.Fatalf("first resume exit=%d: %s", resp.Resume.ExitCode, resp.Resume.Stderr)
	}
	if resp.Resume2.ExitCode != 0 {
		c := resp.Resume2.Stderr + "\n" + resp.Resume2.Stdout
		if strings.Contains(strings.ToLower(c), "already in use") {
			t.Fatalf("second resume already in use:\n%s", c)
		}
		t.Fatalf("second resume exit=%d:\n%s", resp.Resume2.ExitCode, c)
	}
}
```
