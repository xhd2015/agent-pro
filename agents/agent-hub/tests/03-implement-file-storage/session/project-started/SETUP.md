## Preconditions
- A session started event is appended.

## Steps
1. Read the active session projection.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "session-project-started"; return nil }
```

