# Scenario

**Feature**: --set-config and --show-config are mutually exclusive

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--set-config", "--show-config"}
	return nil
}
```
