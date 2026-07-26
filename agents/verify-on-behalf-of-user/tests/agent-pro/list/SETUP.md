# Scenario

**Feature**: agent-pro skills lists verify-on-behalf-of-user

```
agent-pro skills -> line with verify-on-behalf-of-user and description
```

## Steps

1. Invoke `agent-pro skills`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skills"}
	return nil
}
```