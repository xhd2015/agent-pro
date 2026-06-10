## Preconditions
- This branch tests daemon envelopes.

## Steps
1. Mark the test branch as envelope.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    if req.Case == "" { req.Case = "envelope" }
    return nil
}
```

