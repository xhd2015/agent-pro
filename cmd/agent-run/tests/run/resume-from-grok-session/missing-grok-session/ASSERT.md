---
label: e2e
---

## Expected

- Exit code 1.
- Error indicates the Grok session was not found under GROK_HOME.

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
		"not found",
		"no such",
		"unknown session",
		"session not found",
		"missing",
	)
	// Until the flag exists, RED may only say "unknown flag" — still fail
	// the not-found check intentionally (classic TDD).
	_ = strings.ToLower(combined)
}
```
