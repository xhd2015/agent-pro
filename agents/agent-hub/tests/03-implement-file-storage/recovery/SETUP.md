## Preconditions
- This branch tests recovery from the event log.

## Steps
1. Mark the branch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { if req.Case == "" { req.Case = "recovery" }; return nil }
```

