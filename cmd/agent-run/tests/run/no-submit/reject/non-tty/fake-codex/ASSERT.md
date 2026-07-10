## Expected

- Exit code ≠ 0.
- Error clearly indicates `--open` / TTY-only constraint for this runner class
  (not merely “unrecognized flag”).
- Prefer mention of open/TTY/runner constraint wording. Presence of
  `--no-submit` must not bypass the non-TTY gate.

## Errors

- Non-TTY runner with `--open --no-submit`.

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
		t.Fatalf("expected non-zero exit for --open --no-submit + fake-codex; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)

	// Before flags exist, CLI may say "unrecognized flag" — not the product
	// contract for non-TTY rejection under the open family.
	if strings.Contains(errText, "unrecognized flag") || strings.Contains(errText, "unknown flag") {
		t.Fatalf("want non-TTY / open-family rejection, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}

	hasOpenFamily := strings.Contains(errText, "--open") ||
		strings.Contains(errText, "open") ||
		strings.Contains(errText, "--no-submit") ||
		strings.Contains(errText, "no-submit")
	hasTTYWording := strings.Contains(errText, "tty") ||
		strings.Contains(errText, "terminal") ||
		strings.Contains(errText, "non-tty") ||
		strings.Contains(errText, "nontty") ||
		strings.Contains(errText, "interactive") ||
		strings.Contains(errText, "fake-codex") ||
		strings.Contains(errText, "runner")
	if !(hasOpenFamily && hasTTYWording) {
		t.Fatalf("stderr/stdout should explain open family requires a TTY runner:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
