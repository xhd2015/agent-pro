# Scenario

**Feature**: session history -h

```
slack-msg session history -h -> usage
```

## Steps

1. Args: session history -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"session", "history", "-h"}
	return nil
}
```
