# Scenario

**Feature**: listen startup logs config none

```
slack-msg listen --token --app-token -> Using config from: (none)
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
