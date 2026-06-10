## Preconditions
- This branch tests status.

## Steps
1. Mark the branch.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "status" }; return nil }
```

