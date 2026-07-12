## Expected

- Zero sessions / matches returned without error.
- Output is exactly `No sessions found` (same as classic empty list).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 0 {
		t.Fatalf("len(sessions) = %d, want 0", len(resp.Sessions))
	}
	if len(resp.Matches) != 0 {
		t.Fatalf("len(matches) = %d, want 0", len(resp.Matches))
	}
	if strings.TrimSpace(resp.Output) != "No sessions found" {
		t.Fatalf("output = %q, want %q", resp.Output, "No sessions found")
	}
}
```
