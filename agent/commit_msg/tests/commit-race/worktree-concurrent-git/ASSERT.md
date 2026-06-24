## Expected
- `--commit` from a git worktree succeeds without `index.lock` contention.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/git_runner"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("worktree --commit should succeed without index.lock race, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Unable to create") && strings.Contains(resp.Stderr, "index.lock") {
		t.Fatalf("git reported index.lock contention in worktree:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "git commit failed") {
		t.Fatalf("git commit should not fail from worktree, stderr:\n%s", resp.Stderr)
	}

	out, gitErr := git_runner.NewCommand("log", "-1", "--format=%s").Dir(req.GitDir).Output()
	if gitErr != nil {
		t.Fatalf("git log in worktree failed: %v", gitErr)
	}
	subject := strings.TrimSpace(string(out))
	if subject != "feat: worktree commit" {
		t.Fatalf("worktree commit subject = %q, want %q", subject, "feat: worktree commit")
	}
}
```