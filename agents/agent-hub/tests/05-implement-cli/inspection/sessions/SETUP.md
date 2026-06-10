## Preconditions
- A session projection exists.
## Steps
1. List sessions.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "inspection-sessions"; return nil }
```

