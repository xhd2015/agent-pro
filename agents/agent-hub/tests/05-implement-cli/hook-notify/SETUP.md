## Preconditions
- This branch tests hook notify.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "hook-notify" }; return nil }
```

