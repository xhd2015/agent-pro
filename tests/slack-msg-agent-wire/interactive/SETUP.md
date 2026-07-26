# Scenario

**Feature**: runAgentInteractiveOpen cutover to agentrunapi

```
agent.go interactive path
  -> pkgs/agentrunapi + AutoSendOrResume
  -> no RunInteractiveOpen
```

## Steps

1. Mode set per leaf (api vs no-bridge-open).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Leaf overrides Mode to a specific interactive sub-mode.
	if req.Mode == "" {
		req.Mode = "interactive_api"
	}
	return nil
}
```
