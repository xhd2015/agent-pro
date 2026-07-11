## Expected

- Exit code non-zero.
- Stderr mentions that the path is missing / not found / does not exist (case-insensitive).

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for missing --dir path; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	ok := strings.Contains(low, "not found") ||
		strings.Contains(low, "no such") ||
		strings.Contains(low, "does not exist") ||
		strings.Contains(low, "missing") ||
		strings.Contains(low, "cannot find") ||
		strings.Contains(low, "stat")
	if !ok {
		t.Fatalf("stderr should mention missing/not found path; got:\n%s", resp.Stderr)
	}
}
```
