## Preconditions
- Two events exist.
## Steps
1. Fetch without limit.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "fetch-default-limit"; return nil }
```

