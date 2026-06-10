## Preconditions
- One event exists.
## Steps
1. Peek and fetch.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "fetch-peek"; return nil }
```

