# Scenario

**Feature**: session top-level help lists list, info, update, reply, history

```
slack-msg session -h|--help -> lists list / info / update / reply / history; exit 0
```

## Steps

1. Leaf sets help flag after `session`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
