---
label: e2e
---

## Expected Output

```
Weekly limit: 1%
Next reset: July 9, 16:55 PT
```

## Expected

- Exit code 0.
- Stdout contains exactly the two fixture lines (no extra banner noise).
- Stderr is empty.

## Side Effects

- None (ephemeral PTY session only).

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
	assert.Output(t, resp.Stdout, `Weekly limit: 1%
Next reset: July 9, 16:55 PT
`)
}
```