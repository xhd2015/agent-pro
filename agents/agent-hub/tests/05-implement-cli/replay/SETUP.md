## Preconditions
- This branch tests replay.
## Steps
1. Mark branch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "replay" }; return nil }
```

