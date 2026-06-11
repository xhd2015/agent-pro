## Preconditions
- `QUESTION_FIFO` environment variable is NOT set.

## Steps
1. Invoke the binary without `QUESTION_FIFO` in the environment.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{`{"id":"1","question":"test"}`}
    return nil
}
```
