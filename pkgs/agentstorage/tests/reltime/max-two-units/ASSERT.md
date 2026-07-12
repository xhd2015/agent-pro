## Expected

| Age | Output |
|-----|--------|
| 4d5h12m | `4d5h ago` |

- Must not include minutes (`4d5h12m ago` is wrong).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertCases(t, req, resp, err)
	for _, g := range resp.Got {
		if strings.Contains(g, "12m") {
			t.Fatalf("max two units: minutes must be omitted, got %q", g)
		}
	}
}
```
