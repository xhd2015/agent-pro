## Preconditions
- opencode source files exist in all four categories.
- The operation has been set to "export".

## Steps
1. Create `~/.local/share/opencode/auth.json`.
2. Create `~/.local/share/opencode/settings.json`.
3. Create `~/.config/opencode/opencode.jsonc`.
4. Create `~/.config/opencode/plugins/my-plugin.ts`.
5. Create `~/.config/opencode/skills/my-skill/SKILL.md`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "export"
	req.Agent = "opencode"
	createSourceFile(t, req.HomeDir, ".local/share/opencode/auth.json", `{"api_key":"sk-test123"}`)
	createSourceFile(t, req.HomeDir, ".local/share/opencode/settings.json", `{"theme":"dark"}`)
	createSourceFile(t, req.HomeDir, ".config/opencode/opencode.jsonc", `{"command":{"test":{"template":"run tests"}}}`)
	createSourceFile(t, req.HomeDir, ".config/opencode/plugins/my-plugin.ts", `export const MyPlugin = {}`)
	createSourceFile(t, req.HomeDir, ".config/opencode/skills/my-skill/SKILL.md", "# My Skill\n")
	return nil
}
```
