## Expected

- `GlobalSkillDirs()` returns `[$HOME/.codex/skills/, $HOME/.agents/skills/]`

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		filepath.Join(req.Home, ".codex", "skills") + "/",
		filepath.Join(req.Home, ".agents", "skills") + "/",
	}
	if len(resp.Paths) != len(want) {
		t.Fatalf("GlobalSkillDirs() = %d paths, want %d: %v", len(resp.Paths), len(want), resp.Paths)
	}
	for i, w := range want {
		if resp.Paths[i] != w {
			t.Fatalf("GlobalSkillDirs()[%d] = %s, want %s", i, resp.Paths[i], w)
		}
	}
}
```
