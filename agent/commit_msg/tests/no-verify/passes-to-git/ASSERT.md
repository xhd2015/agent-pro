## Expected
- `--commit --no-verify` succeeds despite the failing pre-commit hook.
- A new commit is created with the generated message subject.

## Side Effects
- Git log shows the mock-generated commit title as the latest commit.
- Contrast: the same repo with `--commit` only (no `--no-verify`) would fail at the pre-commit hook.

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
		t.Fatalf("gen-commit-msg --commit --no-verify should succeed past failing hook, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "git commit failed") {
		t.Fatalf("git commit should not fail with --no-verify, stderr:\n%s", resp.Stderr)
	}

	out, gitErr := git_runner.NewCommand("log", "-1", "--format=%s").Dir(req.GitDir).Output()
	if gitErr != nil {
		t.Fatalf("git log failed: %v", gitErr)
	}
	subject := strings.TrimSpace(string(out))
	if subject != "feat: skip hooks" {
		t.Fatalf("commit subject = %q, want %q", subject, "feat: skip hooks")
	}
}
```