# Scenario

**Feature**: session reply help

```
slack-msg session reply -h|--help -> documents --session-id, --config, SLACK_MSG_*
```

## Steps

1. Leaf sets help flag after `session reply`.

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
