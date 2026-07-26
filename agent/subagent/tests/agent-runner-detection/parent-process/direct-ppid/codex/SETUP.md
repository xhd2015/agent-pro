## Preconditions
- `ppid` process name is `"codex"`.

## Steps
1. Set `req.ProcessNames` to `["codex"]`.
2. Priority 4a matches ppid="codex" → returns `"codex"`, `true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"codex"}
    return nil
}
```
