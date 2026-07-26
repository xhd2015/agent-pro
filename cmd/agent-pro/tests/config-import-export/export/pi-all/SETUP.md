## Preconditions
- All pi config files exist under `~/.pi/agent/`.
- The operation has been set to "export".

## Steps
1. Create `~/.pi/agent/auth.json`.
2. Create `~/.pi/agent/settings.json`.
3. Create `~/.pi/agent/models.json`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "export"
	req.Agent = "pi"
	createSourceFile(t, req.HomeDir, ".pi/agent/auth.json", `{"token":"pi-token"}`)
	createSourceFile(t, req.HomeDir, ".pi/agent/settings.json", `{"model":"claude-opus"}`)
	createSourceFile(t, req.HomeDir, ".pi/agent/models.json", `["claude-opus","gpt-4"]`)
	return nil
}
```
