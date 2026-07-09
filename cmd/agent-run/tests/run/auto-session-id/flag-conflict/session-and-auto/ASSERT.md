## Expected

- Exit code ≠ 0.
- Error clearly indicates mutual exclusion / incompatibility of `--session` and
  `--auto-session-id` (not merely "unrecognized flag").
- Both flag names should appear in the error surface, or exclusivity wording
  referencing auto-session-id together with session.

## Errors

- Mutual exclusion between `--session` and `--auto-session-id`.

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
		t.Fatalf("expected non-zero exit when --session and --auto-session-id both set; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)

	// Before the flag exists, CLI may say "unrecognized flag" — that is not the
	// mutual-exclusion contract (still RED).
	if strings.Contains(errText, "unrecognized flag") || strings.Contains(errText, "unknown flag") {
		t.Fatalf("want mutual-exclusion error for --session + --auto-session-id, got unrecognized/unknown flag:\n%s",
			resp.Stderr)
	}

	hasAuto := strings.Contains(errText, "auto-session-id")
	// Require the explicit --session flag token (not the substring inside auto-session-id).
	hasSessionFlag := strings.Contains(errText, "--session")
	hasExclusiveWording := strings.Contains(errText, "mutual") ||
		strings.Contains(errText, "conflict") ||
		strings.Contains(errText, "together") ||
		strings.Contains(errText, "exclusive") ||
		strings.Contains(errText, "incompatible") ||
		strings.Contains(errText, "cannot") ||
		strings.Contains(errText, "not both") ||
		strings.Contains(errText, "either")

	if !(hasAuto && (hasSessionFlag || hasExclusiveWording)) {
		t.Fatalf("stderr/stdout should explain --session vs --auto-session-id mutual exclusion:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
