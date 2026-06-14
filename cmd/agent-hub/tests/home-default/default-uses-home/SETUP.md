## Preconditions
- `AGENT_HUB_HOME` is unset.
- `HOME` points to a temp directory.
- Uses `daemon status` to inspect the resolved home path.

## Steps
1. Run `agent-hub daemon status` without `AGENT_HUB_HOME`.
2. The output JSON `home` field reveals the default path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"daemon", "status"}
    return nil
}
```
