## Expected

- Exit code 1.
- Stderr indicates missing bot token and/or config (clear error).
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
	errOut := strings.ToLower(resp.Stderr)
	if !strings.Contains(errOut, "token") && !strings.Contains(errOut, "config") {
		t.Fatalf("stderr should mention token or config, got:\n%s", resp.Stderr)
	}
}
```
