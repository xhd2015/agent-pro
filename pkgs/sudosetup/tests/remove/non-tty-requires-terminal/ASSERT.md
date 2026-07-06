# Scenario

## Expected

- `Remove` returns an error mentioning interactive terminal / TTY.
- No `sudo rm` runner call.
- Drop-in and manifest remain.

## Side Effects

- None (no sudo attempted).

## Errors

- Error before `sudo rm`.

## Exit Code

- N/A (harness returns error to ASSERT).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, err)
	if hasRunnerCall(resp.RunnerCalls, "sudo", "rm") {
		t.Fatal("sudo rm must not run without interactive stdin")
	}
	if resp.SudoersContent == "" {
		t.Fatal("sudoers drop-in should remain when remove aborts")
	}
	if len(resp.ManifestJSON) == 0 {
		t.Fatal("manifest should remain when remove aborts")
	}
	msg := err.Error()
	if !strings.Contains(msg, "TTY") && !strings.Contains(strings.ToLower(msg), "interactive terminal") {
		t.Fatalf("error = %q, want TTY/interactive terminal hint", msg)
	}
}
```