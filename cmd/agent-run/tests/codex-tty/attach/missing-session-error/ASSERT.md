## Expected

- Exit code 1.
- Stderr mentions session not found or expired (helpful lookup failure).

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "not found") && !strings.Contains(lower, "expired") {
		t.Fatalf("stderr should mention session not found or expired:\n%s", resp.Stderr)
	}
}
```