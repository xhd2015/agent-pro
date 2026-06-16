## Preconditions
- Real opencode config files exist on the host under `~/.local/share/opencode/` or `~/.config/opencode/`.
- Podman is available and the podman machine is running.

## Steps
1. Set the agent to `"opencode"`.
2. Set the query to run the opencode CLI with the danger-zone permissions skip flag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Agent = "opencode"
	req.Query = `opencode run --format json "one word of French capital" --dangerously-skip-permissions`
	return nil
}
```
