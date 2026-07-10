## Expected

- Exit code ≠ 0.
- Error indicates `--open` and `--json` cannot be used together (not merely
  unrecognized flag).
- Both concepts (`open` and `json`) appear in the error surface, or exclusivity
  wording references the combination.

## Errors

- Mutual exclusion between `--open` and `--json`.

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
		t.Fatalf("expected non-zero exit for --open + --json; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)

	if strings.Contains(errText, "unrecognized flag") || strings.Contains(errText, "unknown flag") {
		t.Fatalf("want --open/--json conflict error, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}

	hasOpen := strings.Contains(errText, "open")
	hasJSON := strings.Contains(errText, "json")
	hasExclusive := strings.Contains(errText, "mutual") ||
		strings.Contains(errText, "conflict") ||
		strings.Contains(errText, "together") ||
		strings.Contains(errText, "exclusive") ||
		strings.Contains(errText, "incompatible") ||
		strings.Contains(errText, "cannot") ||
		strings.Contains(errText, "not both") ||
		strings.Contains(errText, "either")

	if !(hasOpen && hasJSON) && !hasExclusive {
		t.Fatalf("stderr/stdout should explain --open vs --json conflict:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	if hasOpen && hasJSON {
		return
	}
	// Exclusive wording alone is weak; still require open|json mention.
	if !(hasOpen || hasJSON) {
		t.Fatalf("stderr/stdout should mention open/json conflict:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
}
```
