## Expected

- `resp.InputBox` is `occupied`.
- Last composer glyph is `»`; that line has no ` medium · `.
- No legacy `›` in the scrollback.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if strings.Contains(req.Scrollback, "›") || strings.Contains(req.Scrollback, "\u203a") {
		t.Fatal("scrollback must not contain ›")
	}
	idx := strings.LastIndex(req.Scrollback, "»")
	if idx < 0 {
		idx = strings.LastIndex(req.Scrollback, "\u00bb")
	}
	if idx < 0 {
		t.Fatal("scrollback must contain »")
	}
	line := req.Scrollback[idx:]
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	if strings.Contains(line, " medium · ") {
		t.Fatalf("last » line must not contain footer glue, got %q", line)
	}
	assertInputBox(t, resp, err, "occupied")
}
```
