## Preconditions
- The envelope contains required metadata and a valid event.

## Steps
1. Select envelope round-trip.

```go
import "testing"

func Setup(t *testing.T, req *Request) error { req.Case = "envelope-valid-round-trip"; return nil }
```

