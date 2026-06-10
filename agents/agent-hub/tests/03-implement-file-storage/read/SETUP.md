## Preconditions
- This branch tests reading event batches.

## Steps
1. Mark the branch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "read" }; return nil }
```

