## Preconditions
- The Pi trace adapter (`agent_trace/pi`) must be imported by the print package.
- Each leaf supplies a Pi event JSON line in `req.Line`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = req
	return nil
}
```
