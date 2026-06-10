## Preconditions
- Two events exist.

## Steps
1. Fetch with limit zero so daemon defaults to one.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "fetch-default-limit-one"; return nil }
```

