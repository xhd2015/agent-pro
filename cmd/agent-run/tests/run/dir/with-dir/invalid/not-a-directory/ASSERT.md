## Expected

- Exit code non-zero.
- Stderr mentions that the path is not a directory (or equivalent).

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
		t.Fatalf("expected non-zero exit for file --dir path; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	ok := strings.Contains(low, "not a directory") ||
		strings.Contains(low, "not a dir") ||
		strings.Contains(low, "is a file") ||
		(strings.Contains(low, "directory") && strings.Contains(low, "not"))
	if !ok {
		t.Fatalf("stderr should mention not a directory; got:\n%s", resp.Stderr)
	}
}
```
