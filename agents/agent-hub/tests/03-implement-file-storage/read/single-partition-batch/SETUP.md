## Preconditions
- Two events exist in one partition.

## Steps
1. Read a two-event batch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "read-single-partition-batch"; return nil }
```

