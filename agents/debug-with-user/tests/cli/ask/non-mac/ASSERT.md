---
label: non-darwin
explanation: asserts macOS-only error path; skipped on darwin hosts where real dialogs are supported
---

## Expected

- Exit code 2 (error).
- Stderr contains a helpful message that macOS/`osascript` is required (or that
  dry-run should be used in CI).
- Stdout does not emit success answer JSON.

## Errors

- Non-macOS without `DEBUG_WITH_USER_DRY_RUN` cannot show real dialogs.

## Exit Code

2

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 2)
	assert.Output(t, resp.Stderr, ``+
`<contains>
macOS
<any-of>
<expect>osascript</expect>
<expect>dry-run</expect>
<expect>DEBUG_WITH_USER_DRY_RUN</expect>
</any-of>
</contains>`)
	trimmed := strings.TrimSpace(resp.Stdout)
	if trimmed != "" && strings.Contains(trimmed, `"via"`) {
		t.Fatalf("error path should not emit answer JSON, got stdout:\n%s", resp.Stdout)
	}
}
```
