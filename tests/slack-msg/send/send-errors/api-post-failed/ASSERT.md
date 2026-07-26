---
label: unit
explanation: slacktest postMessage API error path
---

## Expected

- Exit code 1.
- Stdout may contain Sending line.
- Stderr contains `send failed:`.
- No `OK ts=` in stdout.

## Exit Code

1

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	assertStderrContains(t, resp, "send failed:")
	if strings.Contains(resp.Stdout, "OK ts=") {
		t.Fatalf("unexpected OK line in stdout:\n%s", resp.Stdout)
	}
}
```
