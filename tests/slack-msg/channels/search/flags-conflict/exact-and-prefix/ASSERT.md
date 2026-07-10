## Expected

- Exit code 1.
- Stderr indicates conflict between `--exact` and `--prefix` (mentions both or mutual exclusivity).
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
	hasExact := strings.Contains(low, "exact")
	hasPrefix := strings.Contains(low, "prefix")
	hasMutual := strings.Contains(low, "mutual") || strings.Contains(low, "exclusive") || strings.Contains(low, "together") || strings.Contains(low, "conflict")
	if !(hasExact && hasPrefix) && !hasMutual {
		t.Fatalf("stderr should report --exact/--prefix conflict, got:\n%s", resp.Stderr)
	}
}
```
