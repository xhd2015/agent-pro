---
label: e2e
---

## Expected

- Non-zero exit.
- stderr contains `Error:`.
- stderr mentions `--grep` and empty / non-empty semantics (soft match).
- Fake agent not invoked.

## Side Effects

- None.

## Errors

- Validation error for empty `--grep` pattern.

## Exit Code

- Non-zero.

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
	assertGrepError(t, resp, "--grep")
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")

	lower := strings.ToLower(resp.Stderr)
	// Accept common phrasings: empty / non-empty / must not be empty.
	if !strings.Contains(lower, "empty") && !strings.Contains(lower, "non-empty") {
		t.Fatalf("stderr should mention empty/non-empty pattern semantics:\n%s", resp.Stderr)
	}
}
```
