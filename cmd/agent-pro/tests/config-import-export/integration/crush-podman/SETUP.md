## Preconditions
- Real crush config files exist on the host under `~/.config/crush/` or `~/.local/share/crush/`.
- Podman is available and the podman machine is running.

## Steps
1. Set the agent to `"crush"`.
2. Set the query to run the crush CLI with verbose output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Agent = "crush"
	req.Query = `crush run --verbose "one word of French capital"`
	return nil
}
```
