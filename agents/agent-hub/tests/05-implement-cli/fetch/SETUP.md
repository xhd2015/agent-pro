## Preconditions
- This branch tests fetch.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "fetch" }; return nil }
```

