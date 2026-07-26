---
label: e2e
---

## Expected

- Exit code ≠ 0.
- Error indicates mutual exclusion between `--resume-from-grok-session` and
  `--auto-send-or-resume` (or equivalent "cannot use both" wording).

## Exit Code

≠ 0

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
	assertNonZeroExit(t, resp)
	combined := combinedOut(resp)
	assertContainsAny(t, combined,
		"mutually exclusive",
		"cannot use both",
		"exclusive",
		"conflicts with",
		"incompatible",
	)
	// Anchor to at least one of the two flags so unrelated errors do not pass.
	assertContainsAny(t, combined,
		"resume-from-grok-session",
		"auto-send-or-resume",
	)
	_ = strings.ToLower(combined)
}
```
