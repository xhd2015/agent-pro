# Scenario

**Feature**: bare --set-config without preference flags errors

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--set-config"}
	return nil
}
```
