## Expected

- Exit code 0.
- Stdout includes PTY/limit-related and/or agent-run serve summary fields
  (exact wording is implementer-defined; keywords required).
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if strings.TrimSpace(resp.Stdout) == "" {
		t.Fatalf("pty stats stdout is empty")
	}
	// Require at least one limit/PTY signal and one serve-related signal.
	assertContainsAny(t, resp.Stdout,
		"ptmx", "PTY", "pty", "limit", "Limit", "masters", "free",
	)
	assertContainsAny(t, resp.Stdout,
		"serve", "Serve", "__serve", "agent-run", "orphan", "Holder", "holder",
	)
	assertTrailingNewline(t, resp.Stdout, "pty stats stdout")
}
```
