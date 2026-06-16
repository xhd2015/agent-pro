## Expected
- All non-sensitive files get 0644 permission.
- `settings.json` → 0644
- `models.json` → 0644
- `SKILL.md` → 0644
- `crush/data/crush.json` → 0644

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileMode(t, filepath.Join(req.HomeDir, ".pi/agent/settings.json"), 0644)
	assertFileMode(t, filepath.Join(req.HomeDir, ".pi/agent/models.json"), 0644)
	assertFileMode(t, filepath.Join(req.HomeDir, ".config/opencode/skills/deploy/SKILL.md"), 0644)
	assertFileMode(t, filepath.Join(req.HomeDir, ".local/share/crush/crush.json"), 0644)
}
```
