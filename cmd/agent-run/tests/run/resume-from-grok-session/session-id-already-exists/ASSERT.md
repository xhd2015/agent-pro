---
label: e2e
---

## Expected

- Exit code 1.
- Error indicates the agent-run session id already exists (not "already mapped"
  for the Grok UUID, and not Grok "not found").

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	combined := combinedOut(resp)
	assertContainsAny(t, combined,
		"already exists",
		"already exist",
		"session exists",
		"session already",
		"exists",
	)
	// Must not be the P1 already-mapped path (that uses runner_session_id mapping).
	lower := strings.ToLower(combined)
	if strings.Contains(lower, "already mapped") {
		t.Fatalf("expected session-id collision, not already-mapped:\n%s", combined)
	}
}
```
