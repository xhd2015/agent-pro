## Preconditions
- --session-id is omitted.

## Steps
1. Run `agent-hub session message send --runner fake-opencode --text "hi"` (no --session-id).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    req.Args = []string{"session", "message", "send", "--runner", "fake-opencode", "--text", "hi"}
    return nil
}
```
