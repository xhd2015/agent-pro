## Preconditions
- `ppid` process name is `"zsh"` (zsh shell variant).
- `pppid` process name is `"pi"` (agent runner).

## Steps
1. Set `req.ProcessNames` to `["zsh", "pi"]`.
2. First call (ppid) → "zsh" (no match).
3. Second call (pppid) → "pi" (matches pi-only grandparent walk).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"zsh", "pi"}
    return nil
}
```
