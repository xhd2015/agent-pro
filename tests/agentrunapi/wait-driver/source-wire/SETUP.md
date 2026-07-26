# Scenario

**Feature**: agentruncli wires new-terminal FollowUp to agentrunapi.BuildFollowUpCommand

```
pkgs/agentruncli/*.go
  -> pkgs/agentrunapi import
  -> BuildFollowUpCommand symbol
```

## Steps

1. Set mode `source_wire`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "source_wire"
	return nil
}
```
