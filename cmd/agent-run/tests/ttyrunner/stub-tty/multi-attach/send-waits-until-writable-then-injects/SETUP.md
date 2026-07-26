# Scenario

**Feature**: tty send blocks until writable then injects

```
CheckWritable poll -> ready -> WriteInput
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "send-waits-until-writable-then-injects"
	return nil
}
```
