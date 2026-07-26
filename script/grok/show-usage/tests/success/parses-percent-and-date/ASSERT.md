---
label: e2e
---

## Expected Output

```
Weekly limit: 42%
Next reset: December 25, 12:00 UTC
```

## Expected

- Exit code 0.
- Stdout matches the custom fixture values exactly.
- Stderr is empty.

## Side Effects

- None.

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
	assert.Output(t, resp.Stdout, `Weekly limit: 42%
Next reset: December 25, 12:00 UTC
`)
}
```