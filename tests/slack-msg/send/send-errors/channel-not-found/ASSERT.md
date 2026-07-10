---
label: unit
explanation: conversations.list miss surfaces channel not found
---

## Expected

- Exit code 1.
- Stderr contains `send failed:` and `channel not found`.
- No `OK ts=` in stdout.

## Exit Code

1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStderrContains(t, resp, "send failed:")
	assertStderrContains(t, resp, "channel not found")
	if strings.Contains(resp.Stdout, "OK ts=") {
		t.Fatalf("unexpected OK line in stdout:\n%s", resp.Stdout)
	}
}
```
