## Preconditions
- One daemon already owns the home lock.

## Steps
1. Start a second daemon.

```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "startup-second-daemon-fails"; return nil }
```

