## Preconditions
- `ppid` process name is `"opencode"`.

## Steps
1. Set `req.ProcessNames` to `["opencode"]`.
2. Priority 4a matches ppid="opencode" → returns `"opencode"`, `true`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.ProcessNames = []string{"opencode"}
    return nil
}
```
