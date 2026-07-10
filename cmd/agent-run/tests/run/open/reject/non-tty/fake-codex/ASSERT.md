## Expected

- Exit code ≠ 0.
- Error clearly indicates `--open` is invalid for non-TTY / this runner class
  (not merely “unrecognized flag”).
- Prefer mention of `--open` and TTY / runner constraint wording.

## Errors

- Non-TTY runner with `--open`.

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
		t.Fatalf("expected non-zero exit for --open + fake-codex; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)

	// Before the flag exists, CLI may say "unrecognized flag" — that is not the
	// product contract for non-TTY rejection (still RED until --open lands).
	if strings.Contains(errText, "unrecognized flag") || strings.Contains(errText, "unknown flag") {
		t.Fatalf("want non-TTY / --open rejection, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}

	hasOpen := strings.Contains(errText, "--open") || strings.Contains(errText, "open")
	hasTTYWording := strings.Contains(errText, "tty") ||
		strings.Contains(errText, "terminal") ||
		strings.Contains(errText, "non-tty") ||
		strings.Contains(errText, "nontty") ||
		strings.Contains(errText, "interactive") ||
		strings.Contains(errText, "fake-codex") ||
		strings.Contains(errText, "runner")
	if !(hasOpen && hasTTYWording) {
		t.Fatalf("stderr/stdout should explain --open requires a TTY runner:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
