## Preconditions
- This branch tests shared fake-runner mock config structs.

## Steps
1. Mark the test branch as mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    if req.Case == "" { req.Case = "mock-config" }
    return nil
}
```

