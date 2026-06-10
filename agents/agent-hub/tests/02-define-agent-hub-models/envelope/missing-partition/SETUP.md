## Preconditions
- The envelope omits partition.

## Steps
1. Select missing partition.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "envelope-missing-partition"; return nil }
```

