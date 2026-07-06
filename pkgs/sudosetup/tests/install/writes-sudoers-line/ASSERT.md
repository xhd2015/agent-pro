## Expected

- `EnsureInstalled` succeeds.
- Runner records `sudo visudo -cf` and `sudo install` calls.
- Sudoers drop-in contains `testuser ALL=(root) NOPASSWD:` and command path.

## Side Effects

- Drop-in file written under `SudoersPath`.

## Errors

- None.

## Exit Code

- Success.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !hasRunnerCall(resp.RunnerCalls, "sudo", "visudo") {
		t.Fatal("expected visudo -cf call")
	}
	if !hasRunnerCall(resp.RunnerCalls, "sudo", "install") {
		t.Fatal("expected install call")
	}
	want := "testuser ALL=(root) NOPASSWD: " + req.Rule.Command
	if !strings.Contains(resp.SudoersContent, want) {
		t.Fatalf("sudoers content %q missing %q", resp.SudoersContent, want)
	}
}
```