## Expected

- Exit code 1.
- Stdout empty.
- Stderr reports that `--topic` requires `--help` (e.g. `--topic requires --help`).

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
	if !strings.Contains(low, "--topic") {
		t.Fatalf("stderr should mention --topic, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(low, "--help") && !strings.Contains(low, "help") {
		t.Fatalf("stderr should say --topic requires --help, got:\n%s", resp.Stderr)
	}
	// Prefer the locked form when present.
	if strings.Contains(resp.Stderr, "--topic requires --help") {
		return
	}
	if !strings.Contains(low, "require") {
		t.Fatalf("stderr should indicate --topic requires --help, got:\n%s", resp.Stderr)
	}
}
```
