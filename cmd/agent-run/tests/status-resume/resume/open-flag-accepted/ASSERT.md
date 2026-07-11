## Expected

- Must **not** fail solely because `--open` is an unknown flag.
- Acceptable outcomes:
  - Exit 0 (open resume path completed with instant attach), or
  - Exit ≠ 0 for a **gate/runtime** reason that is **not** unknown-flag.
- Stderr/stdout must not contain "unknown flag", "unknown option", or
  "flag provided but not defined" referring to `--open`.

## Exit Code

0 (preferred) or non-zero only for non-flag reasons

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		// Timeout / hard exec failure is still a fail.
		t.Fatalf("exec error: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	unknownFlagSignals := []string{
		"unknown flag",
		"unknown option",
		"flag provided but not defined",
		"unknown command",
	}
	for _, sig := range unknownFlagSignals {
		if strings.Contains(combined, sig) && strings.Contains(combined, "open") {
			t.Fatalf("--open must be a known resume flag; got unknown-flag style error:\n%s", combined)
		}
	}
	// If exit non-zero, ensure it is not "unknown command: resume" either.
	if resp.ExitCode != 0 {
		if strings.Contains(combined, "unknown command") && strings.Contains(combined, "resume") {
			t.Fatalf("resume command missing: %s", combined)
		}
		// Allowed: runtime issues after flag parse (e.g. attach/TTY env limits).
		t.Logf("resume --open exited %d (flag accepted); stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}
```
