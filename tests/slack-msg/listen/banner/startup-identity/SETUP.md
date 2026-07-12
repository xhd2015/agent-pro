# Scenario

**Feature**: banner includes bot/team identity and listen settings

```
listen connects -> banner lines include team, bot user/id, modes, lock path
```

## Steps

1. Inherit banner grouping (daemon, explicit lock file).
2. Assert multi-field banner on combined stdout/stderr (no events required).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Identity banner only needs connected daemon; no inbound events.
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
