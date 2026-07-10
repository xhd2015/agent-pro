## Expected

- Exit code 1.
- Stderr mentions query required (contains `query` and `required`, or `QUERY`).
- Stdout empty.

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
	if resp.Stdout != "" {
		t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "query") || !strings.Contains(low, "required") {
		t.Fatalf("stderr should report query required, got:\n%s", resp.Stderr)
	}
}
```
