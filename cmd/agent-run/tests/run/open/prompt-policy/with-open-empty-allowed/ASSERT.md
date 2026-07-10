## Expected

- Must not fail with `prompt is required`.
- Ideally exit 0 once `--open` is implemented (instant attach + keep-alive path).
- If the process errors for other reasons (e.g. attach hook missing mid-impl),
  the error surface still must not be the empty-prompt validation message.

## Errors

- None for empty-prompt validation.

## Exit Code

0 preferred (non-zero allowed only if message is not `prompt is required`)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	// Timeout / exec failure still fails the leaf (not "prompt is required").
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(errText, "prompt is required") {
		t.Fatalf("--open with empty prompt must not report prompt is required:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	// Full GREEN contract: success after open lifecycle.
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --open without prompt once implemented; exit=%d\nstderr:\n%s\nstdout:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}
```
