## Expected

- Attach capture contains live session output (`ATTACH_LIVE_MARKER` or boot marker).
- Attacher receives multiplexed PTY bytes (not empty).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected detached session id")
	}
	out := resp.AttachOutput
	if out == "" {
		out = resp.Combined
	}
	if out == "" {
		t.Fatal("attach produced no output from detached session")
	}
	if !strings.Contains(out, "ATTACH_LIVE_MARKER") && !strings.Contains(out, "ATTACH_BOOT_MARKER") {
		t.Fatalf("attach missing live session markers, got %q", out)
	}
}
```