# Scenario

**Feature**: session history validation errors

```
missing session id | unknown session -> stderr; exit 1
```

## Steps

1. Clear env; isolate home.
2. Leaf sets invalid args.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	return nil
}
```
