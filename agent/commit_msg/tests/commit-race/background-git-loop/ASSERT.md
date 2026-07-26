## Expected
- `--commit` completes successfully even when the agent left a stale `index.lock` behind.
- stderr must not report git `Unable to create` / `index.lock` errors.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/git_runner"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("gen-commit-msg --commit should succeed without index.lock race, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Unable to create") && strings.Contains(resp.Stderr, "index.lock") {
		t.Fatalf("git reported index.lock contention:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "git commit failed") {
		t.Fatalf("git commit should not fail, stderr:\n%s", resp.Stderr)
	}

	out, gitErr := git_runner.NewCommand("log", "-1", "--format=%s").Dir(req.GitDir).Output()
	if gitErr != nil {
		t.Fatalf("git log failed: %v", gitErr)
	}
	subject := strings.TrimSpace(string(out))
	if subject != "feat: session ID auto-generation" {
		t.Fatalf("commit subject = %q, want %q", subject, "feat: session ID auto-generation")
	}
}
```