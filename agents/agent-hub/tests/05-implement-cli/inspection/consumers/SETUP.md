## Preconditions
- A consumer cursor exists.
## Steps
1. List consumers.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "inspection-consumers"; return nil }
```

