---
label: e2e
---

## Expected Output

```
Monthly usage: 58%
Credits used: 6519 of 11250
Next reset: 08:00 on 1 Aug
```

## Expected

- Exit code 0.
- Stdout matches the default three canonical lines.
- Stderr is empty.
- `codex-status-usage` registry file is **removed** after CLI exits.

## Side Effects

- tty-watch session killed; registry entry pruned.

## Errors

- None.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessExit(t, resp)
	assert.Output(t, resp.Stdout, `Monthly usage: 58%
Credits used: 6519 of 11250
Next reset: 08:00 on 1 Aug
`)
	assertRegistrySessionGone(t, req.TTYWatchHome, req.SessionID)
}
```