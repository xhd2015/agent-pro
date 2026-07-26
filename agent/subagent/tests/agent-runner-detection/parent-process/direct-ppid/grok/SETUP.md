## Preconditions
- `ppid` process name is `"grok"`.

## Steps
1. Set `req.ProcessNames` to `["grok"]`.
2. Priority 4a matches ppid="grok" → returns `"grok"`, `true` (no grandparent walk).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"grok"}
    return nil
}
```
