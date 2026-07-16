## Expected

- Must not fail with `prompt is required`.
- Exit code 0 once `--detach` is implemented.
- Stdout prints both `session-id:` and `terminal-id:`.

## Errors

- None for empty-prompt validation.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(errText, "prompt is required") {
		t.Fatalf("--detach with empty prompt must not report prompt is required:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	assertNotUnknownFlag(t, errText, "--detach")
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --detach without prompt once implemented; exit=%d\nstderr:\n%s\nstdout:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertDetachIDsOnStdout(t, resp)
}
```
