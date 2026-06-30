## Expected

- Exit code 1.
- Stderr explains that `--print` is required or usage is shown.

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(combined, "print") && !strings.Contains(combined, "usage") {
		t.Fatalf("expected stderr/stdout to mention --print or usage, got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```