## Preconditions
- No seed is provided; a random seed is generated internally.

## Steps
1. Run fake opencode without a mock config or seed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"run", "--format", "json", "hello world"}
    return nil
}
```
