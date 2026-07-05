## Expected

- `LocalSkillDirs(projectDir)` returns `[projectDir/.opencode/skills/, projectDir/.claude/skills/, projectDir/.agents/skills/]`

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		filepath.Join(req.ProjectDir, ".opencode", "skills") + "/",
		filepath.Join(req.ProjectDir, ".claude", "skills") + "/",
		filepath.Join(req.ProjectDir, ".agents", "skills") + "/",
	}
	if len(resp.Paths) != len(want) {
		t.Fatalf("LocalSkillDirs() = %d paths, want %d: %v", len(resp.Paths), len(want), resp.Paths)
	}
	for i, w := range want {
		if resp.Paths[i] != w {
			t.Fatalf("LocalSkillDirs()[%d] = %s, want %s", i, resp.Paths[i], w)
		}
	}
}
```
