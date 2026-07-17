# Scenario

**Feature**: agent-run sources call BuildFollowUpCommand for new-terminal path

```
scan cmd/agent-run/*.go for BuildFollowUpCommand + agentrunapi import
```

## Steps

1. Mode already `source_wire`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Mode != "source_wire" {
		req.Mode = "source_wire"
	}
	return nil
}
```
