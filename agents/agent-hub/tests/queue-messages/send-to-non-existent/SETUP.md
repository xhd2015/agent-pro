## Preconditions
- No prior session exists.

## Steps
1. Send message to a non-existent session ID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "AGENT_HUB_OPENCODE_RUNNER=fake-opencode")
    return nil
}
```
