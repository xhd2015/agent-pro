## Preconditions
- `ppid` process name is `"bash"` (common shell).
- `pppid` process name is `"grok"` (agent runner).

## Steps
1. Set `req.ProcessNames` to `["bash", "grok"]`.
2. First call (ppid) → "bash" (no match).
3. Second call (pppid) → "grok" (matches pi+grok grandparent walk).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"bash", "grok"}
    return nil
}
```
