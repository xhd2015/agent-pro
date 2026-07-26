---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Error clearly indicates `--no-submit` requires `--open` (not merely
  “unrecognized flag”).
- Both concepts (`no-submit` / no submit, and `open`) appear in the error
  surface, or exclusivity/requirement wording references the pair.

## Errors

- `--no-submit` without `--open`.

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --no-submit without --open; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)

	// Before the flag exists, CLI may say "unrecognized flag" — that is not the
	// product contract for --no-submit requires --open (still RED until landed).
	if strings.Contains(errText, "unrecognized flag") || strings.Contains(errText, "unknown flag") {
		t.Fatalf("want --no-submit requires --open error, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}

	hasNoSubmit := strings.Contains(errText, "--no-submit") ||
		strings.Contains(errText, "no-submit") ||
		strings.Contains(errText, "nosubmit") ||
		(strings.Contains(errText, "no") && strings.Contains(errText, "submit"))
	hasOpen := strings.Contains(errText, "--open") || strings.Contains(errText, "open")
	hasRequires := strings.Contains(errText, "require") ||
		strings.Contains(errText, "must") ||
		strings.Contains(errText, "need") ||
		strings.Contains(errText, "only with") ||
		strings.Contains(errText, "together") ||
		strings.Contains(errText, "without")

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
