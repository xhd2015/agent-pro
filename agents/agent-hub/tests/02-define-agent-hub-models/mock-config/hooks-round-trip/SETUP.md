## Preconditions
- The mock config contains one hook.

## Steps
1. Select hook round-trip scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "mock-config-hooks-round-trip"; return nil }
```

