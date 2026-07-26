# Scenario

**Feature**: DEBUG_WITH_USER_DRY_RUN simulates user choices without osascript

```
DEBUG_WITH_USER_DRY_RUN=1 + staged env -> ask -> JSON answer on stdout
```

## Steps

1. Set `DEBUG_WITH_USER_DRY_RUN=1` on every dry-run leaf.
2. Leaf setup adds button/text/dismissed staging env vars.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Env = append(req.Env, "DEBUG_WITH_USER_DRY_RUN=1")
	return nil
}
```
