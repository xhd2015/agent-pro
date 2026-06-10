## Preconditions
- Session projection was deleted before restart.

## Steps
1. Restart daemon.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "restart-rebuilds-sessions"; return nil }
```

