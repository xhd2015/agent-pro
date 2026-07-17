# Scenario

**Feature**: cmd/agent-run wires new-terminal FollowUp to agentrunapi.BuildFollowUpCommand

```
cmd/agent-run/*.go
  -> pkgs/agentrunapi import
  -> BuildFollowUpCommand symbol
```

## Steps

1. Set mode `source_wire`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "source_wire"
	return nil
}
```
