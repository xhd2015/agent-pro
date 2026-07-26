## Preconditions
- The test runs `fake-opencode models`.

## Steps
1. Run the models command.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"models"}
    return nil
}
```

