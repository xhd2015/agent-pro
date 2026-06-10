## Preconditions
- This branch tests restart.

## Steps
1. Mark the branch.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "restart" }; return nil }
```

