# Scenario

**Feature**: ForceNew / `BuildFollowUpCommand` site forwards Model + Effort

```
pkgs/agentruncli openAutoInNewTerminal (or equivalent)
  -> BuildFollowUpCommand(FollowUpOpts{ Model, ModelReasoningEffort, … })
```

## Steps

1. Set mode `source_wire`.
2. Leaf sets `SourceWireTarget=follow_up_forward`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "source_wire"
	return nil
}
```
