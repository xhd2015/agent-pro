## Expected

| Age | Output | Not |
|-----|--------|-----|
| 1h0m5s | `1h ago` | `1h5s ago` |
| 4d0h2m | `4d ago` | `4d2m ago` |

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertCases(t, req, resp, err)
	// Guard against the wrong “skip zeros” implementation.
	for _, g := range resp.Got {
		if strings.Contains(g, "1h5s") || strings.Contains(g, "4d2m") {
			t.Fatalf("zero unit must stop the chain, got %q", g)
		}
	}
}
```
