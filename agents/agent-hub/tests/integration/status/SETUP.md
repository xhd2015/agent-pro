## Preconditions
- Status reports plugin presence and enabled/disabled state.

## Steps
1. Run `agent-hub integration status` with various states.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    _ = t
    return nil
}
```
