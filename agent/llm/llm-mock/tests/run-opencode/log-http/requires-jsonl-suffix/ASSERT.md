---
label: e2e
---

## Expected

- Non-zero exit code.
- Combined stderr/stdout mentions `.jsonl`.
- Log file at the bad path was not created.
- Opencode did not run (`OPENCODE_RAN` absent; orchestrator did not reach `OPENCODE_CONFIG_DIR=` announcement).

## Errors

- Must fail validation before starting mock server and opencode.

## Exit Code

non-zero

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, ".jsonl")
	assertNotContains(t, combined, "OPENCODE_RAN")
	assertNotContains(t, combined, "OPENCODE_CONFIG_DIR=")

	if _, statErr := os.Stat(req.LogHTTPPath); statErr == nil {
		t.Fatalf("log file should not exist at %q", req.LogHTTPPath)
	}
}
```