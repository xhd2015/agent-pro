## Preconditions
- A cursor is saved for consumer `c1`.

## Steps
1. Load the saved cursor.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "cursor-save-load"; return nil }
```

