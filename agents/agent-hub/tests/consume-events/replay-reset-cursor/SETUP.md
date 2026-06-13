## Preconditions
- 5 events produced, consumer has fetched 3 (cursor advanced).

## Steps
1. Produce 5 events.
2. Fetch first 3 events.
3. Replay --from partition:0.
4. Fetch --limit 5 to verify all events returned.

```go
import (
    "testing"
    "fmt"
)

func Setup(t *testing.T, req *Request) error {
    for i := 0; i < 5; i++ {
        notifyEvent(t, req, "agent.session.started", "fake-opencode", fmt.Sprintf("s_replay_%d", i))
    }
    return nil
}
```
