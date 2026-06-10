## Preconditions
- This branch tests hook normalization errors.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "errors" }; return nil }
```

