---
label: e2e
---

## Expected

- Exit code 1.
- Error indicates mutual exclusion between auto-send-or-resume and
  `--session-id-from-prompt` (and/or requires explicit session-id).

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
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	// Prefer explicit mutex wording; fall back to session-id-from-prompt + error.
	assertContainsAny(t, combined,
		"mutually exclusive",
		"cannot use both",
		"session-id-from-prompt",
	)
	assertContainsAny(t, combined,
		"auto-send-or-resume",
		"session-id-from-prompt",
		"session-id",
	)
}
```
