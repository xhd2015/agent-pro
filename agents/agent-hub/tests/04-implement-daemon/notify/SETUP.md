## Preconditions
- This branch tests notify.

## Steps
1. Mark the branch.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "notify" }; return nil }
```

