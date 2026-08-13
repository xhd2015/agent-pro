## Expected

- API error mentions JSON.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	if !strings.Contains(strings.ToLower(resp.ErrString), "json") {
		t.Fatalf("error should mention json, got %q", resp.ErrString)
	}
}
```
