## Expected
- Exactly two `turn_completed` wire lines.

```go
import (
	"testing"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range resp.WireLines {
		if strings.Contains(line, `"sessionUpdate":"turn_completed"`) {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("turn_completed count: got %d want 2\n%s", n, resp.Output)
	}
}
```
