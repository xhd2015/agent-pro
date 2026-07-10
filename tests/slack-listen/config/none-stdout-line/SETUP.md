# Scenario

**Feature**: startup logs config none

```
slack-listen listen --bot-token --app-token -> Using config from: (none)
```

## Steps

1. Daemon probe with CLI tokens only (no --config).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Daemon = true
	req.WantAgentCalls = 0
	return nil
}
```