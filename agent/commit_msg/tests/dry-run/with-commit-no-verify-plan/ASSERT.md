## Expected
- Dry-run with `--commit --no-verify` succeeds.
- Stdout is mock B for N=1.
- Stderr would-line includes both `git commit` and `--no-verify`.
- HEAD subject is unchanged.

## Side Effects
- No commit is created; hooks are never run.
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
		t.Fatalf("dry-run --commit --no-verify should succeed as plan, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	AssertMockMessageB(t, resp.Stdout, 1)
	AssertNoAgentInvoked(t, resp)

	if !strings.Contains(strings.ToLower(resp.Stderr), "would:") ||
		!strings.Contains(resp.Stderr, "git commit") {
		t.Fatalf("stderr should contain would: git commit plan, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "--no-verify") {
		t.Fatalf("would-line should include --no-verify, stderr:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", resp.Stderr)
	}

	subject := GitHEADSubject(t, req.GitDir)
	if subject != req.Operation {
		t.Fatalf("HEAD subject changed: before=%q after=%q", req.Operation, subject)
	}
}
```
