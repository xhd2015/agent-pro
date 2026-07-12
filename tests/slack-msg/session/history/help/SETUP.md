# Scenario

**Feature**: session history help

```
slack-msg session history -h|--help -> documents --session-id, --after-msg-id, --json
```

## Steps

1. Leaf sets help flag after `session history`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
