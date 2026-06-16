## Preconditions
- `ppid` process name is `"zsh"` (alternate shell).
- `pppid` process name is `"grok"` (agent runner).

## Steps
1. Set `req.ProcessNames` to `["zsh", "grok"]`.
2. First call (ppid) → "zsh" (no match).
3. Second call (pppid) → "grok" (matches pi+grok grandparent walk).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ProcessNames = []string{"zsh", "grok"}
    return nil
}
```
