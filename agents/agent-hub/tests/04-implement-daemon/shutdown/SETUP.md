## Preconditions
- This branch tests shutdown.

## Steps
1. Mark the branch.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "shutdown" }; return nil }
```

