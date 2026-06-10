## Preconditions
- Valid event JSON is passed inline.
## Steps
1. Notify event.
```go
import "testing"
func Setup(t *testing.T, req *Request) error { req.Case = "notify-json-valid"; return nil }
```

