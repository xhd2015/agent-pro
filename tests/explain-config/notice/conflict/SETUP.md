# Scenario

**Feature**: --color and --no-color together hard-error

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--color", "--no-color", "hello"}
	return nil
}
```
