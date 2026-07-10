---
label: unit
explanation: channel resolution only; fake token send fails
---

## Expected

- Exit code 1.
- Stdout contains `Sending to channel=C0ALE44K5J6: "Hello slack"`.
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
	assertStdoutContains(t, resp, `Sending to channel=C0ALE44K5J6: "Hello slack"`)
	assertStderrContains(t, resp, "send failed:")
	if strings.Contains(resp.Stdout, "OK ts=") {
		t.Fatalf("unexpected success OK line in stdout:\n%s", resp.Stdout)
	}
}
```