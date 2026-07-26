# Scenario

**Feature**: session reply -h

```
slack-msg session reply -h -> usage
```

## Steps

1. Args: session reply -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"session", "reply", "-h"}
	return nil
}
```
