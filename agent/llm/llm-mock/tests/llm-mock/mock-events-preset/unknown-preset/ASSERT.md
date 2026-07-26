---
label: e2e
---

## Expected

- Non-zero exit code (server did not announce a listening port).
- Combined stderr/stdout mentions the unknown preset name or preset resolution error.
- No HTTP responses collected.

## Errors

- Must fail validation before the server listens for HTTP traffic.

## Exit Code

non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.ExitCode == 0 && resp.Err == nil {
		t.Fatalf("expected startup failure for unknown preset, got exit 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "nonexistent")

	if resp.Port != 0 {
		t.Fatalf("unknown preset must not start server; got port %d", resp.Port)
	}
	if len(resp.Responses) != 0 {
		t.Fatalf("expected no HTTP responses, got %d", len(resp.Responses))
	}
}
```