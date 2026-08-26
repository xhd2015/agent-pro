# Scenario

**Feature**: --show-config with missing file prints {}

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--show-config"}
	return nil
}
```
