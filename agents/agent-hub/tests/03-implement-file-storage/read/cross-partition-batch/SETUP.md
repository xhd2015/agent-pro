## Preconditions
- Events exist in two adjacent partitions.

## Steps
1. Read a batch that crosses partitions.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "read-cross-partition-batch"; return nil }
```

