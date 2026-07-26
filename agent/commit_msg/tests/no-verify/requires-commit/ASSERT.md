## Expected
- gen-commit-msg exits with an error before calling the agent.
- Error message states that `--no-verify` requires `--commit`.

## Side Effects
- No commit is created.
- stderr must not contain agent-phase logs (`Passing diff to agent`, `Running agent`).

## Exit Code
- Non-zero.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err == nil {
		t.Fatalf("expected gen-commit-msg to fail with --no-verify alone, stderr:\n%s", resp.Stderr)
	}
	errMsg := resp.Err.Error()
	if !strings.Contains(errMsg, "--no-verify") || !strings.Contains(errMsg, "--commit") {
		t.Fatalf("error should mention --no-verify requires --commit, got: %v", resp.Err)
	}
	for _, marker := range []string{"Passing diff to agent", "Running agent"} {
		if strings.Contains(resp.Stderr, marker) {
			t.Fatalf("agent should not run before flag validation, found %q in stderr:\n%s", marker, resp.Stderr)
		}
	}
}
```