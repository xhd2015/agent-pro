## Preconditions
- This branch tests consumer cursors.

## Steps
1. Mark the branch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "cursor" }; return nil }
```

