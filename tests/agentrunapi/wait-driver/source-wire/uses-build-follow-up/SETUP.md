# Scenario

**Feature**: agentruncli sources call BuildFollowUpCommand for new-terminal path

```
scan pkgs/agentruncli/*.go for BuildFollowUpCommand + agentrunapi import
```

## Steps

1. Mode already `source_wire`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "source_wire" {
		req.Mode = "source_wire"
	}
	return nil
}
```
