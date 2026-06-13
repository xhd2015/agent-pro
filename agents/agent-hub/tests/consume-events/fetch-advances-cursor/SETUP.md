## Preconditions
- 3 events produced.

## Steps
1. Produce 3 events.
2. First fetch --limit 1.
3. Second fetch --limit 1.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    for i := 0; i < 3; i++ {
        notifyEvent(t, req, "agent.session.started", "fake-opencode", "s_adv")
    }
    return nil
}
```
