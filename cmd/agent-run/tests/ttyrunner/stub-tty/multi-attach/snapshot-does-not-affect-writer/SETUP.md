# Scenario

**Feature**: snapshot probe does not claim write token

```
snapshot attach -> ephemeral; writer retains unified write
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "snapshot-does-not-affect-writer"
	return nil
}
```
