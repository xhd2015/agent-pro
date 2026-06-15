## Preconditions
- Grouping node for FromPi conversion tests.

## Steps
1. Sets `Target` to `"from_pi"` as a default for all FromPi leaves.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_pi"
	return nil
}
```
