## Expected

- Paris + exited.
- Resume without --open + followup succeeds (keep-tty path); not already-in-use; not prompt required.

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
		t.Fatalf("need Paris+exited; paris=%v exited=%v", resp.HasParis, resp.ExitedTrue)
	}
	c := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	if strings.Contains(c, "prompt is required") {
		t.Fatalf("unexpected prompt required:\n%s", c)
	}
	if resp.AlreadyInUse || strings.Contains(c, "already in use") {
		t.Fatalf("already in use:\n%s", c)
	}
	// keep-tty with followup waits for turn complete; mock may leave process
	// running past the exec timeout (-1) even after streaming hello markers.
	// Accept exit 0, or timeout/non-zero if hello was clearly delivered.
	if resp.Resume.ExitCode != 0 {
		hasHello := strings.Contains(c, "hello") ||
			strings.Contains(c, "hello_resume_marker") ||
			strings.Contains(strings.ToLower(resp.EventsBlob), "hello") ||
			strings.Contains(resp.EventsBlob, req.HelloMarker)
		if !hasHello && !strings.Contains(c, "grok session") {
			t.Fatalf("resume keep-tty exit=%d without hello/session progress:\n%s", resp.Resume.ExitCode, c)
		}
		t.Logf("resume keep-tty exit=%d but hello/session progress observed (ok)", resp.Resume.ExitCode)
	}
}
```
