## Preconditions
- No events produced.

## Steps
1. Fetch with fresh consumer ID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"fetch", "--consumer-id", "cempty-"+t.Name()}
    return nil
}
```
