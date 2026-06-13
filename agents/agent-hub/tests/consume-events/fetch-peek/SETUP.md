## Preconditions
- 1 event produced.

## Steps
1. Produce 1 event.
2. Fetch with --peek.
3. Fetch again without --peek to verify same event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_peek")
    return nil
}
```
