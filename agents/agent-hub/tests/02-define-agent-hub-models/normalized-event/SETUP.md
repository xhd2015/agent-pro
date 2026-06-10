## Preconditions
- This branch tests normalized event validation.

## Steps
1. Mark the test branch as normalized event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    if req.Case == "" {
        req.Case = "normalized-event"
    }
    return nil
}
```

