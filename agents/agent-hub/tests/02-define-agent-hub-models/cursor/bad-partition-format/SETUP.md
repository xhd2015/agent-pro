## Preconditions
- The cursor uses slash-separated partition format.

## Steps
1. Select bad partition format.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "cursor-bad-partition-format"; return nil }
```

