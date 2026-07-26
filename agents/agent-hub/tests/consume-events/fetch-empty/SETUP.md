## Preconditions
- No events produced.

## Steps
1. Fetch with fresh consumer ID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"fetch", "--consumer-id", "cempty-"+t.Name()}
    return nil
}
```
