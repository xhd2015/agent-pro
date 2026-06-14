## Preconditions
- `AGENT_HUB_HOME` is set to a custom temp directory.
- Verifies the env var takes precedence over the default.

## Steps
1. Set `AGENT_HUB_HOME` to a custom path.
2. Run `agent-hub daemon status`.
3. Confirm `home` matches the custom path.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    customHome := filepath.Join(req.TempDir, "custom-agent-hub")
    req.Env = append(req.Env, "AGENT_HUB_HOME="+customHome)
    req.Env = append(req.Env, "HOME="+os.Getenv("HOME"))
    req.Args = []string{"daemon", "status"}
    return nil
}
```
