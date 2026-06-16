## Preconditions
- Real pi config files exist on the host under `~/.pi/agent/`.
- Podman is available and the podman machine is running.

## Steps
1. Set the agent to `"pi"`.
2. Set the query to run the pi CLI with JSON mode and auto-approve.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Agent = "pi"
	req.Query = `pi -p "one word of French capital" --mode json --approve`
	return nil
}
```
