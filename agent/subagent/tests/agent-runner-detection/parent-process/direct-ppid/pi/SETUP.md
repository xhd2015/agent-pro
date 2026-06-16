## Preconditions
- `ppid` process name is `"pi"`.

## Steps
1. Set `req.ProcessNames` to `["pi"]`.
2. Priority 4a matches ppid="pi" → returns `"pi"`, `true` (no grandparent walk).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ProcessNames = []string{"pi"}
    return nil
}
```
