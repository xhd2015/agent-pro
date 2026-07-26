## Preconditions
- Both crush config files exist.
- The operation has been set to "export".

## Steps
1. Create `~/.config/crush/crush.json` (global config with provider keys).
2. Create `~/.local/share/crush/crush.json` (data store).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "export"
	req.Agent = "crush"
	createSourceFile(t, req.HomeDir, ".config/crush/crush.json", `{"provider":"anthropic","api_key":"sk-crush"}`)
	createSourceFile(t, req.HomeDir, ".local/share/crush/crush.json", `{"projects":[]}`)
	return nil
}
```
