## Expected

- `resp.InputBox` is `empty`.
- TrimSpace of the remainder after last `❯` is empty.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	idx := strings.LastIndex(req.Scrollback, "❯")
	if idx < 0 {
		t.Fatal("scrollback must contain ❯")
	}
	line := req.Scrollback[idx+len("❯"):]
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	if strings.TrimSpace(line) != "" {
		t.Fatalf("padding-only remainder required, got %q", line)
	}
	assertInputBox(t, resp, err, "empty")
}
```
