---
label: e2e
---

## Expected

- Non-zero exit code.
- Combined stderr/stdout mentions `.jsonl`.
- Log file at the bad path was not created.
- Grok did not run (`GROK_RAN` absent; orchestrator did not reach `GROK_HOME=` announcement).

## Errors

- Must fail validation before starting mock server and grok.

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
	assertNotContains(t, combined, "GROK_RAN")
	assertNotContains(t, combined, "GROK_HOME=")

	if _, statErr := os.Stat(req.LogEventsPath); statErr == nil {
		t.Fatalf("log file should not exist at %q", req.LogEventsPath)
	}
}
```