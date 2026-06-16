## Preconditions
- `ppid` process name is `"bash"` (common shell).
- `pppid` process name is `"pi"` (agent runner).

## Steps
1. Set `req.ProcessNames` to `["bash", "pi"]`.
2. First call (ppid) → "bash" (no match).
3. Second call (pppid) → "pi" (matches pi-only grandparent walk).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ProcessNames = []string{"bash", "pi"}
    return nil
}
```
