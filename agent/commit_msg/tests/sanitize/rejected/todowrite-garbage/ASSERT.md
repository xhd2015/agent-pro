## Expected
- gen-commit-msg returns a non-nil error (unusable message after sanitize).
- stdout must not be a plausible clean commit message containing the garbage as subject.
- With `--commit`, HEAD is unchanged.

## Errors
- Error or stderr should indicate the message is unusable (substring from `.want_err` optional match).

## Exit Code
- Non-zero (resp.Err set / ExitCode != 0).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	if resp.Err == nil {
		t.Fatalf("expected hard failure on todowrite garbage, got success\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	// HEAD must not move when --commit was requested.
	AssertHEADUnchanged(t, req.GitDir, req.WorktreeDir)

	wantSub := ReadAntiPatternWantErr(t, "todowrite_garbage")
	errBlob := resp.Err.Error() + "\n" + resp.Stderr
	// Prefer matching .want_err substring when implementer lands a reason string;
	// until then, any hard failure without a new commit is the contract.
	if wantSub != "" && !strings.Contains(strings.ToLower(errBlob), strings.ToLower(wantSub)) {
		// Soft note: do not fail solely on reason wording if error is present —
		// but require non-success. Log for implementer visibility.
		t.Logf("error did not contain want_err substring %q (ok if reason text differs): %v", wantSub, resp.Err)
	}
	if strings.Contains(resp.Stdout, "todowrite") {
		t.Fatalf("garbage leaked to stdout:\n%s", resp.Stdout)
	}
}
```
