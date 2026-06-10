## Preconditions
- An index file is removed after append.

## Steps
1. Rebuild indexes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "index-rebuild-missing"; return nil }
```

