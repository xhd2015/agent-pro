## Preconditions
- The cursor starts on a day without a partition.

## Steps
1. Read the next available partition.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "read-missing-day-skipped"; return nil }
```

