---
label: e2e, real-codex, slow
explanation: Requires real codex CLI; first-turn tool-bash preset must execute mocked bash.
---

## Expected

- Exit code 0.
- Combined stdout/stderr contains `preset-bash` (output of mocked `echo preset-bash` bash tool on first turn).

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
	assertContains(t, combined, "preset-bash")
}
```