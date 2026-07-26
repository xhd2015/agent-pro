## Preconditions
- `ppid` process name is `"crush"`.

## Steps
1. Set `req.ProcessNames` to `["crush"]`.
2. Priority 4a matches ppid="crush" → returns `"crush"`, `true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.ProcessNames = []string{"crush"}
    return nil
}
```
