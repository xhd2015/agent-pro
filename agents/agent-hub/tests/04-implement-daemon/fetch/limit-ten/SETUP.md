## Preconditions
- Three events exist.

## Steps
1. Fetch with limit ten.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "fetch-limit-ten"; return nil }
```

