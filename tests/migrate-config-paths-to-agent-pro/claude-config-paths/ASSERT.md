## Expected

- `SettingsPath()` is `$HOME/.claude/settings.json`
- `JSONConfigPath()` is `$HOME/.claude.json`
- `GlobalSkillsDir()` is `$HOME/.claude/skills/`

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
	if len(resp.Paths) != 3 {
		t.Fatalf("expected 3 paths, got %d: %v", len(resp.Paths), resp.Paths)
	}

	want := []string{
		filepath.Join(req.Home, ".claude", "settings.json"),
		filepath.Join(req.Home, ".claude.json"),
		filepath.Join(req.Home, ".claude", "skills") + "/",
	}
	for i, w := range want {
		if resp.Paths[i] != w {
			t.Fatalf("path[%d] = %s, want %s", i, resp.Paths[i], w)
		}
	}
}
```
