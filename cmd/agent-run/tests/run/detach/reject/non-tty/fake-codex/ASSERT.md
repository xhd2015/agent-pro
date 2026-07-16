## Expected

- Exit code ≠ 0.
- Error clearly indicates `--detach` is invalid for non-TTY / this runner class
  (not merely “unrecognized flag”).
- Prefer product-style wording:
  `--detach requires a TTY runner (got …)`

## Errors

- Non-TTY runner with `--detach`.

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
		t.Fatalf("expected non-zero exit for --detach + fake-codex; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertNotUnknownFlag(t, errText, "--detach")

	hasDetach := strings.Contains(errText, "--detach") || strings.Contains(errText, "detach")
	hasTTYWording := strings.Contains(errText, "tty") ||
		strings.Contains(errText, "terminal") ||
		strings.Contains(errText, "non-tty") ||
		strings.Contains(errText, "nontty") ||
		strings.Contains(errText, "interactive") ||
		strings.Contains(errText, "fake-codex") ||
		strings.Contains(errText, "runner")
	if !(hasDetach && hasTTYWording) {
		t.Fatalf("stderr/stdout should explain --detach requires a TTY runner:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
