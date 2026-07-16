## Expected
- gen-commit-msg fails with a non-nil error.
- Error message contains `not a git repository`.

## Side Effects
- Agent is not invoked.
- No commit is created (no git repo).

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
		t.Fatalf("expected not-a-git-repository error, stdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	errMsg := resp.Err.Error()
	if !strings.Contains(errMsg, "not a git repository") {
		t.Fatalf("error should contain %q, got: %v", "not a git repository", resp.Err)
	}
	for _, marker := range []string{"Passing diff to agent", "Running agent"} {
		if strings.Contains(resp.Stderr, marker) {
			t.Fatalf("agent should not run before IsInsideGit failure, found %q in stderr:\n%s", marker, resp.Stderr)
		}
	}
}
```
