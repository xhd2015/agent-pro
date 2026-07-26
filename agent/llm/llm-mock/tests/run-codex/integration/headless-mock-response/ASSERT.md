---
label: e2e, real-codex, slow
explanation: Requires real codex CLI on PATH; headless LLM round-trip up to 60s.
---

## Expected

- Exit code 0.
- Combined stdout/stderr contains `"Paris"` (mocked LLM response).

## Side Effects

- Codex session transcript may be written under `$CODEX_HOME/sessions/...` (not asserted in MVP).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "Paris")
}
```