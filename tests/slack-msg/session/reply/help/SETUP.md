# Scenario

**Feature**: session reply help

```
slack-msg session reply -h|--help -> documents --session-id, --config, SLACK_MSG_*
```

## Steps

1. Leaf sets help flag after `session reply`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
