## Preconditions
- This branch tests session projection.

## Steps
1. Mark the branch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "session" }; return nil }
```

