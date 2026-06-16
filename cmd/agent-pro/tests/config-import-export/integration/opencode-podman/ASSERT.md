## Expected
- The opencode CLI output contains the word "paris" (case-insensitive).
- The agent responds using the configs that were exported from the host and imported into the container.
- No timeout or execution errors.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("opencode query failed: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if !strings.Contains(strings.ToLower(resp.Output), "paris") {
		t.Fatalf("expected output to contain 'paris', got: %q", resp.Output)
	}
}
```
