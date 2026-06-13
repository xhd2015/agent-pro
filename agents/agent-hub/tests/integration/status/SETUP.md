## Preconditions
- Status reports plugin presence and enabled/disabled state.

## Steps
1. Run `agent-hub integration status` with various states.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    _ = t
    return nil
}
```
