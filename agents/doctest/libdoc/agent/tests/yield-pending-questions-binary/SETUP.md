## Preconditions
- Tests in this group run the `yield-pending-questions` binary directly.
- The binary is dispatched via `os.Args[0]` from the doctest binary.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.AsYieldPQ = true
    return nil
}
```
