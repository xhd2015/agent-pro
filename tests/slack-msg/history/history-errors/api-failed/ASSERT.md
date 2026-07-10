---
label: unit
explanation: conversations.history API error path
---

## Expected

- Exit code 1.
- Stderr contains `history failed:`.

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStderrContains(t, resp, "history failed:")
}
```
