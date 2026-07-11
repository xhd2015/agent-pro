## Expected

- Exit code 1.
- Stderr (or stdout) indicates cannot resume / runner not exited / use `send`.

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
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"cannot resume",
		"not exited",
		"still active",
		"use send",
		"use `send`",
		"agent-run send",
		"exited",
	)
	// Prefer a send hint when gate is live-not-exited.
	assertContainsAny(t, combined, "send", "live", "active", "running")
}
```
