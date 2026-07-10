---
label: unit
explanation: conversations.list API error path
---

## Expected

- Exit code 1.
- Stderr contains `channels failed:`.

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStderrContains(t, resp, "channels failed:")
}
```
