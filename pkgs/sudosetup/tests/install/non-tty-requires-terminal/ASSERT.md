# Scenario

## Expected

- `EnsureInstalled` returns an error mentioning interactive terminal / TTY.
- No `visudo` or `install` runner calls.
- No manifest written.

## Side Effects

- None (no sudo attempted).

## Errors

- Error before any `sudo` subprocess.

## Exit Code

- N/A (harness returns error to ASSERT).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, err)
	if hasRunnerCall(resp.RunnerCalls, "sudo", "visudo") {
		t.Fatal("visudo must not run without interactive stdin")
	}
	if hasRunnerCall(resp.RunnerCalls, "sudo", "install") {
		t.Fatal("install must not run without interactive stdin")
	}
	if len(resp.ManifestJSON) != 0 {
		t.Fatal("manifest must not be written without interactive stdin")
	}
	msg := err.Error()
	if !strings.Contains(msg, "TTY") && !strings.Contains(strings.ToLower(msg), "interactive terminal") {
		t.Fatalf("error = %q, want TTY/interactive terminal hint", msg)
	}
}
```