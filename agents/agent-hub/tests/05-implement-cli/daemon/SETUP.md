## Preconditions
- This branch tests daemon CLI commands.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "daemon" }; return nil }
```

