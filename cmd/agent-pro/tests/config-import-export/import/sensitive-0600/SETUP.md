## Preconditions
- Files that may contain secrets are written with restricted permissions.
- The operation has been set to "import".

## Steps
1. Create a zip with sensitive files: auth.json, opencode.jsonc, crush/config/crush.json.
2. Run import.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "import"
	req.Agent = "opencode"
	createZip(t, req.ZipPath, map[string]string{
		"pi/auth.json":               `{"token":"secret"}`,
		"opencode/opencode.jsonc":    `{"permission":{"bash":"ask"}}`,
		"crush/config/crush.json":    `{"api_key":"sk-secret"}`,
		"opencode/settings.json":     `{"theme":"dark"}`,
	})
	return nil
}
```
