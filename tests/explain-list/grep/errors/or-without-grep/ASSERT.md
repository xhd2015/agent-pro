---
label: e2e
---

## Expected

- Non-zero exit.
- stderr contains `Error:`.
- stderr mentions `--grep` (and preferably `--or`).
- Fake agent not invoked.

## Side Effects

- None.

## Errors

- Mode flag without greps.

## Exit Code

- Non-zero.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertGrepError(t, resp, "--grep")
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")
}
```
