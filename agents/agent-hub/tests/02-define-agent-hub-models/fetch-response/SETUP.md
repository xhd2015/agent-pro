## Preconditions
- This branch tests fetch response JSON.

## Steps
1. Mark the test branch as fetch response.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    if req.Case == "" { req.Case = "fetch-response" }
    return nil
}
```

