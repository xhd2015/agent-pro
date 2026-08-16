---
label: unit
explanation: session history human lines oldest→newest from local log
---

## Expected Output

```
[1710000901.000100] U1: first
[1710000902.000200] U2: second
[1710000903.000300] U1: third
```

## Expected

- Exit code 0.
- Three human lines oldest→newest; trailing newline.
- Stderr empty.

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
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `[1710000901.000100] U1: first
[1710000902.000200] U2: second
[1710000903.000300] U1: third
`)
}
```
