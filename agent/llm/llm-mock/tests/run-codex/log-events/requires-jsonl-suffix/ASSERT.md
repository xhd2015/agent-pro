## Expected

- Non-zero exit code.
- Combined stderr/stdout mentions `.jsonl`.
- Log file at the bad path was not created.
- Codex did not run (`CODEX_RAN` absent; orchestrator did not reach `CODEX_HOME=` announcement).

## Errors

- Must fail validation before starting mock server and codex.

## Exit Code

non-zero

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, ".jsonl")
	assertNotContains(t, combined, "CODEX_RAN")
	assertNotContains(t, combined, "CODEX_HOME=")

	if _, statErr := os.Stat(req.LogEventsPath); statErr == nil {
		t.Fatalf("log file should not exist at %q", req.LogEventsPath)
	}
}
```