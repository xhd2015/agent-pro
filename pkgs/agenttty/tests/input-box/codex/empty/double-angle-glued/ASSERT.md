## Expected

- `resp.InputBox` is `empty`.
- Scrollback contains `»` and ` medium · ` on that line; no legacy `›`.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if !strings.Contains(req.Scrollback, "»") && !strings.Contains(req.Scrollback, "\u00bb") {
		t.Fatal("scrollback must contain »")
	}
	if strings.Contains(req.Scrollback, "›") || strings.Contains(req.Scrollback, "\u203a") {
		t.Fatal("scrollback must not contain › (would mask the » path)")
	}
	if !strings.Contains(req.Scrollback, " medium · ") {
		t.Fatal("scrollback must contain footer glue")
	}
	assertInputBox(t, resp, err, "empty")
}
```
