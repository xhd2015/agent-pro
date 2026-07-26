# Scenario

**Feature**: explicit config logging at listen startup

```
slack-msg listen -> logs Using config from: (none) or absolute --config path
```

## Preconditions

- No auto-discovery of slack-config.json.

## Steps

1. Leaf chooses no config, explicit config, or bad path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	return nil
}
```
