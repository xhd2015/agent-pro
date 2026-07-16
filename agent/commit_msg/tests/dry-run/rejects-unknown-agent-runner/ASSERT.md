## Expected
- gen-commit-msg returns an error for unsupported agent runner under `--dry-run`.
- Error mentions unsupported runner / `codex` / supported `opencode` (existing pattern).

## Side Effects
- Mock success path is not taken (stdout is not mock B alone as success).
- Agent is not invoked.

## Exit Code
- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err == nil {
		t.Fatalf("expected unsupported agent runner error under dry-run, stdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	errMsg := resp.Err.Error()
	if !strings.Contains(errMsg, "unsupported agent runner") {
		t.Fatalf("error should use unsupported agent runner pattern, got: %v", resp.Err)
	}
	if !strings.Contains(errMsg, "codex") {
		t.Fatalf("error should mention codex, got: %v", resp.Err)
	}
	if !strings.Contains(errMsg, "opencode") {
		t.Fatalf("error should mention supported opencode, got: %v", resp.Err)
	}
	AssertNoAgentInvoked(t, resp)
}
```
