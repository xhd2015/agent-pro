# Scenario

**Feature**: explicit config logging at startup

```
slack-listen listen -> logs Using config from: (none) or absolute --config path
```

## Preconditions

- No auto-discovery of slack-config.json.

## Steps

1. Leaf chooses no config, explicit config, or bad path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	return nil
}
```