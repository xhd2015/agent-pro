## Expected

- First commit under immutable index fails; output contains `unable to write new index file`.
- That first failure is classified as transient by `IsTransientIndexError`.
- After clearing uchg, `CommitWithRetry` succeeds.
- HEAD subject equals `feat: recovered after index write fail`.

## Side Effects

- New commit on temp repo after recovery.
- Index flags restored to writable (`nouchg`) before retry.

## Errors

- First single `Commit` must error; final `CommitWithRetry` must not.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.FirstFailErr == nil {
		t.Fatalf("expected first commit under uchg to fail, got success:\n%s", resp.FirstFailOutput)
	}
	if !strings.Contains(resp.FirstFailOutput, "unable to write new index file") {
		t.Fatalf("first failure should contain production message, got:\n%s", resp.FirstFailOutput)
	}
	if !resp.FirstFailTransient {
		t.Fatalf("first failure must be classified transient:\n%s", resp.FirstFailOutput)
	}
	if resp.CommitErr != nil {
		t.Fatalf("CommitWithRetry after clear should succeed: %v\noutput:\n%s", resp.CommitErr, resp.Output)
	}
	if resp.Subject != req.Message {
		t.Fatalf("commit subject = %q, want %q", resp.Subject, req.Message)
	}
}
```
