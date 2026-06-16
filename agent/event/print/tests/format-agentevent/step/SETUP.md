## Preconditions
- The event Type is either `step_start` or `step_finish`.

## Steps
- Each leaf sets the specific step type.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = req
	return nil
}
```
