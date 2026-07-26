---
label: e2e, ui-automation
explanation: Requires playwright-debug and browser automation.
---

## Expected

- Playwright exits 0.
- Finished chat page shows a visible enabled terminal button.

## Side Effects

- Browser localStorage stores the test auth token.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until the frontend uses terminal availability
  independently from `finished` status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s", resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
