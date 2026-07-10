---
label: unit
explanation: daemon startup log probe
---

## Expected

- Startup output contains `Using config from: (none)`.
- Process stops cleanly on SIGTERM.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertStdoutContains(t, resp, "Using config from: (none)")
}
```