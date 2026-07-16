## Expected

- Exit code ≠ 0 (parse failure, validation, or usage error).
- Error indicates an empty / missing id for `--resume-from-grok-session`
  (not merely that the flag is unknown — classic TDD stays RED until the flag
  is registered and empty-id is rejected).

## Exit Code

≠ 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNonZeroExit(t, resp)
	combined := strings.TrimSpace(combinedOut(resp))
	if combined == "" {
		t.Fatal("expected error output for empty --resume-from-grok-session")
	}
	// Require empty-value validation wording so "unrecognized flag" alone stays RED.
	assertContainsAny(t, combined,
		"empty",
		"non-empty",
		"requires a non-empty",
		"require a non-empty",
		"must not be empty",
		"cannot be empty",
	)
}
```
