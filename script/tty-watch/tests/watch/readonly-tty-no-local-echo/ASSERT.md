## Expected

- Typed characters sent to the watch PTY must not appear in captured output
  (no local terminal echo).
- Mouse-report escape sequences must not appear in captured output.
- Remote session output (`WATCH_MARKER`) may still stream.
- Observer mode is readonly; input is silently dropped.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if strings.Contains(combined, "WATCH_LOCAL_ECHO_PROBE") {
		t.Fatalf("watch locally echoed typed input on TTY, got %q", combined)
	}
	if strings.Contains(combined, "[<0;") || strings.Contains(combined, "\x1b[<0;") {
		t.Fatalf("watch locally echoed mouse input on TTY, got %q", combined)
	}
}
```