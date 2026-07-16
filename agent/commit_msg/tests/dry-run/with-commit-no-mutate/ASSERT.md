## Expected
- Dry-run with `--commit` succeeds (plan only).
- Stdout is mock B for N=1.
- Stderr contains a would-commit line (`would: git commit`).
- HEAD subject is unchanged from before the run.

## Side Effects
- No new commit is created.
- Agent is not invoked.

## Exit Code
- Zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("dry-run --commit should succeed without committing, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	AssertMockMessageB(t, resp.Stdout, 1)
	AssertNoAgentInvoked(t, resp)

	if !strings.Contains(strings.ToLower(resp.Stderr), "would:") ||
		!strings.Contains(resp.Stderr, "git commit") {
		t.Fatalf("stderr should contain would: git commit plan, stderr:\n%s", resp.Stderr)
	}
	// Real commit path markers must not appear as an executed commit.
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", resp.Stderr)
	}

	subject := GitHEADSubject(t, req.GitDir)
	if subject != req.Operation {
		t.Fatalf("HEAD subject changed under dry-run --commit: before=%q after=%q", req.Operation, subject)
	}
}
```
