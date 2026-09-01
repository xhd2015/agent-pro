## Expected

- List returns zero models without error.
- Human output contains `(no models)`.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Catalog.Models) != 0 {
		t.Fatalf("Models=%v want empty", resp.Catalog.Models)
	}
	if !strings.Contains(resp.Output, "(no models)") {
		t.Fatalf("output missing (no models):\n%s", resp.Output)
	}
}
```
