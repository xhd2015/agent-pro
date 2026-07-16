## Expected
- gen-commit-msg fails with an error about no staged changes.
- Error text matches the existing generate path wording (contains `no staged`).

## Side Effects
- Agent is not invoked.
- HEAD is unchanged.

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
		t.Fatalf("expected no-staged error under dry-run, stdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	errMsg := strings.ToLower(resp.Err.Error())
	if !strings.Contains(errMsg, "no staged") {
		t.Fatalf("error should mention no staged changes, got: %v", resp.Err)
	}
	AssertNoAgentInvoked(t, resp)
}
```
