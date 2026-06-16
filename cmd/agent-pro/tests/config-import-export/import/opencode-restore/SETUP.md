## Preconditions
- A valid zip file with opencode entries exists at the zip path.
- The operation has been set to "import".

## Steps
1. Create a zip file containing opencode config entries.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "import"
	req.Agent = "opencode"
	createZip(t, req.ZipPath, map[string]string{
		"opencode/auth.json":                    `{"api_key":"sk-exported"}`,
		"opencode/settings.json":                `{"theme":"light"}`,
		"opencode/opencode.jsonc":               `{"command":{"test":{"template":"run"}}}`,
		"opencode/plugins/hook.ts":              `export const Hook = {}`,
		"opencode/skills/deploy/SKILL.md":       "# Deploy skill\n",
	})
	return nil
}
```
