## Expected

- `CommitWithRetry` returns a non-nil commit error (hook blocked the commit).
- Combined output is **not** classified as transient (`IsTransientIndexError` false).
- No successful commit with the requested subject.

## Side Effects

- HEAD remains the seed commit (no new commit from this attempt).

## Errors

- Expected: git / hook failure from `CommitWithRetry`.
- `Run` itself returns nil error; failure is in `resp.CommitErr`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/git_runner"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.CommitErr == nil {
		t.Fatalf("CommitWithRetry must fail on pre-commit hook, but succeeded; subject=%q output:\n%s", resp.Subject, resp.Output)
	}
	if git_runner.IsTransientIndexError(resp.Output, resp.CommitErr) {
		t.Fatalf("hook failure must not be transient; would cause pointless retries:\n%s\nerr: %v", resp.Output, resp.CommitErr)
	}
	// Subject should still be seed if commit did not land
	if resp.Subject == req.Message {
		t.Fatalf("should not have committed with message %q after hook failure", req.Message)
	}
	lower := strings.ToLower(resp.Output)
	if !strings.Contains(lower, "hook") && !strings.Contains(lower, "pre-commit") {
		// Soft check: git usually mentions the hook; still require fail-fast above.
		t.Logf("hook-related wording not found in output (still failed as required):\n%s", resp.Output)
	}
}
```
