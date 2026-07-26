## Expected
- stdout is the inner title from the first `-m` argument (no residual `git commit -m` wrapper form).
- The cleaned subject may still contain the words "git commit" as ordinary title text.

## Expected Output

```
---
version: 2
---
feat: extract message from git commit -m wrapper
```

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("gen-commit-msg should extract git -m title, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	want := ReadAntiPatternWant(t, "git_commit_m_wrapper")
	AssertStdoutMessage(t, resp.Stdout, want)
	assert.Output(t, resp.Stdout, `---
version: 2
---
feat: extract message from git commit -m wrapper
`)
	// Reject residual wrapper form, not the words inside a clean subject.
	trimmed := strings.TrimSpace(resp.Stdout)
	if strings.HasPrefix(trimmed, "`git") || strings.HasPrefix(trimmed, "git commit -m") {
		t.Fatalf("git commit wrapper leaked into stdout:\n%s", resp.Stdout)
	}
}
```
