# Scenario

**Feature**: ForceNew follow-up builder assigns Model and ModelReasoningEffort

```
scan pkgs/agentruncli files that call BuildFollowUpCommand
  -> composite literal includes Model: and ModelReasoningEffort:
```

## Steps

1. Target `follow_up_forward`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "source_wire"
	req.SourceWireTarget = "follow_up_forward"
	return nil
}
```
