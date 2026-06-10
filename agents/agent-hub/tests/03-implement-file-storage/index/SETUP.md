## Preconditions
- This branch tests index rebuild.

## Steps
1. Mark the branch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "index" }; return nil }
```

