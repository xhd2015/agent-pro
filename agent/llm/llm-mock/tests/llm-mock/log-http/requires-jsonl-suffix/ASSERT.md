---
label: e2e
---

## Expected

- Non-zero exit code (server did not announce a listening port).
- Combined stderr/stdout mentions `.jsonl`.
- Log file at the bad path was not created.

## Errors

- Must fail validation before the server listens for HTTP traffic.

## Exit Code

non-zero

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.ExitCode == 0 && resp.Err == nil {
		t.Fatalf("expected startup failure, got exit 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, ".jsonl")

	if _, statErr := os.Stat(req.LogHTTPFile); statErr == nil {
		t.Fatalf("log file should not exist at %q", req.LogHTTPFile)
	}
}
```