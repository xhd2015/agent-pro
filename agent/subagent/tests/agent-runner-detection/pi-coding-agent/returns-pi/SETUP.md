## Preconditions
- `PI_CODING_AGENT` env var is set to `"true"`.

## Steps
1. Set `PI_CODING_AGENT=true` in env.
2. Priority 3 matches → returns `"pi"`, `true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = []string{"PI_CODING_AGENT=true"}
    return nil
}
```
