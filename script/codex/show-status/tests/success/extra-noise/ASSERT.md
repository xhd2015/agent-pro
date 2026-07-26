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
- Stdout contains exactly the three canonical lines despite MCP noise in scrollback.
- Stderr is empty.

## Side Effects

- Ephemeral tty-watch session killed after fetch.

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
}
```