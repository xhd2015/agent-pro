# Scenario

**Feature**: tty send times out after 10s with provider reason

```
busy screen never writable -> exit 1 with reason after 10s
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "send-times-out-with-reason-after-10s"
	return nil
}
```
