## Expected
- gen-commit-msg succeeds without `auto unstage failed` when `--dir` is a subdirectory of the git repo root.
- Staged repo-root-relative paths such as `task-hub/agents/do-task/PROMPT.md` are resolved against the repository root, not the nested working directory.
- The generated commit message is printed to stdout.

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
		if strings.Contains(resp.Err.Error(), "auto unstage failed") {
			t.Fatalf("auto unstage should resolve repo-root paths from a nested --dir, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
		}
		t.Fatalf("gen-commit-msg failed: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "docs: update do-task prompt") {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}
}
```