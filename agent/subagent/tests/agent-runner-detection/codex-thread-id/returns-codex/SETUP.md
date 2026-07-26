## Preconditions
- `CODEX_THREAD_ID` env var is set to `"abc123"`.

## Steps
1. Set `CODEX_THREAD_ID=abc123` in env.
2. Priority 2 matches → returns `"codex"`, `true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = []string{"CODEX_THREAD_ID=abc123"}
    return nil
}
```
