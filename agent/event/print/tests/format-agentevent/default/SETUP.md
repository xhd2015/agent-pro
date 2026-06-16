## Preconditions
- The event has an unhandled ActionType that falls through to the default
  `[type] + text` formatting.

## Steps
- Each leaf sets a specific non-standard type.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = req
	return nil
}
```
