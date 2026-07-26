---
label: e2e, ui-automation
explanation: Playwright NotFound page + home link
---

## Expected

- Playwright exits 0.
- `[data-testid="not-found"]` visible on unknown client route.
- Click `[data-testid="not-found-home"]` navigates to `/`.
- Nav marker survives (client Link / soft nav).

## Side Effects

- Background web cleanup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
