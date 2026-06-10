## Preconditions
- Session projection files are deleted.

## Steps
1. Rebuild session projection from the event log.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "recovery-rebuild-sessions"; return nil }
```

