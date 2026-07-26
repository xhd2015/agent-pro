## Expected
- gen-commit-msg succeeds.
- stdout is the sanitized formatted message (outer backticks removed from title).
- `git log -1 --format=%s` is the cleaned title only.

## Side Effects
- A new commit is created with the cleaned subject (not the backticked title).

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
	if resp.Err != nil {
		t.Fatalf("gen-commit-msg should succeed after sanitize, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "title_outer_backticks")
	AssertStdoutMessage(t, resp.Stdout, want)
	title := strings.SplitN(want, "\n", 2)[0]
	if strings.Contains(resp.Stdout, "`feat:") || strings.Contains(resp.Message, "`") {
		t.Fatalf("outer backticks leaked into output:\n%s", resp.Stdout)
	}
	AssertGitSubject(t, req.GitDir, title)
}
```
