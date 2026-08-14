## Expected

- `resp.InputBox` is `empty` (last glyph is the glued empty line).
- Scrollback contains an earlier `› leftover` **and** a later ` medium · ` on the last `›` line.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if !strings.Contains(req.Scrollback, "› leftover") && !strings.Contains(req.Scrollback, "\u203a leftover") {
		t.Fatal("scrollback must include historical leftover draft")
	}
	idx := strings.LastIndex(req.Scrollback, "›")
	if idx < 0 {
		idx = strings.LastIndex(req.Scrollback, "\u203a")
	}
	if idx < 0 {
		t.Fatal("scrollback must contain ›")
	}
	line := req.Scrollback[idx:]
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	if !strings.Contains(line, " medium · ") {
		t.Fatalf("last › line must be empty-glued, got %q", line)
	}
	assertInputBox(t, resp, err, "empty")
}
```
