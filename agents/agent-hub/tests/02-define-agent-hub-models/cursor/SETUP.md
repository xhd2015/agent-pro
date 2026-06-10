## Preconditions
- This branch tests cursor validation.

## Steps
1. Mark the test branch as cursor.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    if req.Case == "" { req.Case = "cursor" }
    return nil
}
```

