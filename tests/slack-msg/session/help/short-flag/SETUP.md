# Scenario

**Feature**: session -h

```
slack-msg session -h -> usage lists reply and history
```

## Steps

1. Args: session -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"session", "-h"}
	return nil
}
```
