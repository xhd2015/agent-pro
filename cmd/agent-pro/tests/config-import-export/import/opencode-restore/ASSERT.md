## Expected
- All opencode files are restored to their correct disk locations.
- `auth.json` → `~/.local/share/opencode/auth.json`
- `settings.json` → `~/.local/share/opencode/settings.json`
- `opencode.jsonc` → `~/.config/opencode/opencode.jsonc`
- `plugins/hook.ts` → `~/.config/opencode/plugins/hook.ts`
- `skills/deploy/SKILL.md` → `~/.config/opencode/skills/deploy/SKILL.md`
- File contents match the original zip entries.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileExists(t, filepath.Join(req.HomeDir, ".local/share/opencode/auth.json"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".local/share/opencode/settings.json"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".config/opencode/opencode.jsonc"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".config/opencode/plugins/hook.ts"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".config/opencode/skills/deploy/SKILL.md"))
}
```
