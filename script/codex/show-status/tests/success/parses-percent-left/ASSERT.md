## Expected Output

```
Monthly usage: 70%
Credits used: 3000 of 10000
Next reset: 12:00 on 15 Jan
```

## Expected

- Exit code 0.
- Stdout matches the custom fixture values exactly (`30% left` → `70%` usage).
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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessExit(t, resp)
	assert.Output(t, resp.Stdout, `Monthly usage: 70%
Credits used: 3000 of 10000
Next reset: 12:00 on 15 Jan
`)
}
```