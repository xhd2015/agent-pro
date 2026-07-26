---
label: e2e
---

## Expected

- Exit code 0.
- Combined stdout/stderr contains `MODELS_CACHE=mock-model-found` (orchestrator wrote `models_cache.json` with slug `mock-model`).

## Side Effects

- `$CODEX_HOME/models_cache.json` exists before codex child starts.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "MODELS_CACHE=mock-model-found")
}
```