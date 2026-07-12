## Expected

- Exit code ≠ 0.
- Error indicates `--no-submit` requires `--open`.
- Must not be an unrecognized/unknown-flag error alone for `--no-submit`.

## Errors

- `--no-submit` without `--open` on resume.

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
		t.Fatalf("expected non-zero exit for resume --no-submit without --open; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(errText, "unrecognized flag") ||
		(strings.Contains(errText, "unknown flag") && strings.Contains(errText, "no-submit")) {
		t.Fatalf("want --no-submit requires --open error, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}
	hasNoSubmit := strings.Contains(errText, "--no-submit") ||
		strings.Contains(errText, "no-submit") ||
		(strings.Contains(errText, "no") && strings.Contains(errText, "submit"))
	hasOpen := strings.Contains(errText, "--open") || strings.Contains(errText, "open")
	hasRequires := strings.Contains(errText, "require") ||
		strings.Contains(errText, "must") ||
		strings.Contains(errText, "need")
	if !(hasNoSubmit && hasOpen) {
		t.Fatalf("stderr/stdout should explain --no-submit requires --open:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	if !hasRequires {
		t.Fatalf("stderr/stdout should use requires/must wording for --no-submit + --open:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
