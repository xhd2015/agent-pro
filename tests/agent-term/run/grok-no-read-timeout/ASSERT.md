## Expected

- After 4 seconds attached to `grok`, PTY output must not contain `i/o timeout`.
- Grok should have started (non-empty output or grok terminal mode sequences).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(resp.Combined, "i/o timeout") {
		t.Fatalf("grok attach hit websocket read timeout: %q", resp.Combined)
	}
	if strings.TrimSpace(resp.Combined) == "" {
		t.Fatalf("expected grok attach output, got empty")
	}
}
```