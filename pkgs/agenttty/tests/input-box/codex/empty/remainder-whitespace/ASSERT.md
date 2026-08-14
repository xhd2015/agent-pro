## Expected

- `resp.InputBox` is `empty`.
- TrimSpace of the remainder after `›` is empty.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(req.Scrollback), "›"))
	if rest != "" {
		t.Fatalf("setup remainder must be whitespace-only, got %q", rest)
	}
	assertInputBox(t, resp, err, "empty")
}
```
