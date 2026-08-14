## Expected

- `resp.InputBox` is `occupied` (Grok ignores Codex ` medium · ` glue).
- Last `❯` remainder is non-empty after TrimSpace.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if !strings.Contains(req.Scrollback, " medium · ") {
		t.Fatal("scrollback must include Codex-looking footer glue")
	}
	if strings.ContainsAny(req.Scrollback, "›»") {
		t.Fatal("this leaf must be Grok-only (no Codex glyph)")
	}
	assertInputBox(t, resp, err, "occupied")
}
```
