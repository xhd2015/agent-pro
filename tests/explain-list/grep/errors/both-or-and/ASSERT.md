---
label: e2e
---

## Expected

- Non-zero exit.
- stderr contains `Error:`.
- stderr mentions both `--or` and `--and` (or clear mutual-exclusion wording).
- Fake agent not invoked.

## Side Effects

- None.

## Errors

- Conflicting combine flags.

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
	assertGrepError(t, resp, "--or", "--and")
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")

	// Soft lock: both flags named, or explicit conflict language.
	lower := strings.ToLower(resp.Stderr)
	if !(strings.Contains(lower, "--or") && strings.Contains(lower, "--and")) {
		t.Fatalf("stderr should mention both --or and --and:\n%s", resp.Stderr)
	}
}
```
