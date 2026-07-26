## Expected
- Dry-run with `--model some/model` succeeds (model is accepted, not used).
- Stdout is mock B for N=1.
- Agent is not invoked.

## Side Effects
- No commit; index unchanged for the staged file.

## Exit Code
- Zero.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("dry-run --model should succeed without agent, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	AssertMockMessageB(t, resp.Stdout, 1)
	AssertNoAgentInvoked(t, resp)

	staged := GitStagedNames(t, req.GitDir)
	if len(staged) == 0 {
		t.Fatalf("expected staged files to remain, got empty")
	}
}
```
