## Preconditions
- Fetch limit is zero.
## Steps
1. Fetch with invalid limit.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "fetch-limit-zero-invalid"; return nil }
```

