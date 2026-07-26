---
label: e2e, real-codex, slow
explanation: Requires real codex CLI on PATH; reproduces model-metadata warning on mock-model.
---

## Expected

- Exit code 0.
- Combined stdout/stderr contains mocked preset message text `preset:message:think-tool-message`.
- Combined stdout/stderr does **not** contain `Model metadata for \`mock-model\` not found`.

## Side Effects

- Codex may write session transcript under `$CODEX_HOME/sessions/...` (not asserted).

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
	assertContains(t, combined, "preset:message:think-tool-message")
	assertNotContains(t, combined, "Model metadata for `mock-model` not found")
}
```