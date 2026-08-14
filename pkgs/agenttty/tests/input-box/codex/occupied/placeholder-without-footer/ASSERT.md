## Expected

- `resp.InputBox` is `occupied`.
- Scrollback has no ` medium · ` (so this is not the live-empty shape).

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if strings.Contains(req.Scrollback, " medium · ") {
		t.Fatal("this leaf must not include footer glue")
	}
	assertInputBox(t, resp, err, "occupied")
}
```
