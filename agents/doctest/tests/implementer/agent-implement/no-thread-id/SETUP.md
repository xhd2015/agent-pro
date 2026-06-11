## Preconditions
- `CODEX_THREAD_ID` is not set in the environment.

## Steps
1. Run `doctest agent implement` without `CODEX_THREAD_ID`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"agent", "implement", "some prompt"}
    return nil
}
```
