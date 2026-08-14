## Expected

- `resp.InputBox` is `unknown`.
- Scrollback has no `›` / `»` / `❯`.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if strings.ContainsAny(req.Scrollback, "›»❯") {
		t.Fatal("fixture must not contain a composer glyph")
	}
	assertInputBox(t, resp, err, "unknown")
}
```
