---
label: unit
explanation: channel resolution only; fake token send fails
---

## Expected

- Exit code 1.
- Stdout contains `Sending to channel=G12345ABCDE: "Hello slack"`.
- Stderr contains `send failed:`.

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStdoutContains(t, resp, `Sending to channel=G12345ABCDE: "Hello slack"`)
	assertStderrContains(t, resp, "send failed:")
}
```