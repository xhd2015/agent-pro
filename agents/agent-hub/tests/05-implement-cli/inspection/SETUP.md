## Preconditions
- This branch tests inspection commands.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "inspection" }; return nil }
```

