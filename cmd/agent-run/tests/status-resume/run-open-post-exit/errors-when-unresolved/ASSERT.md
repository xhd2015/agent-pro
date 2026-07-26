---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Stderr contains an error about grok session id not resolved (for the session).
- May still print `grok-tty: <id>` terminal line before the error.

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
		t.Fatalf("exec error: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when grok session unresolved; stderr:\n%s\nstdout:\n%s", resp.Stderr, resp.Stdout)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"grok session id not resolved",
		"session id not resolved",
		"not resolved for session",
		"not resolved",
	)
	// Prefer the product-facing phrasing from the requirement when present.
	if !strings.Contains(combined, "not resolved") {
		t.Fatalf("missing not-resolved wording:\n%s", combined)
	}
}
```
