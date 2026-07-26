# Scenario

**Feature**: cmd session JSONL events are converted to/from canonical AgentEvents

```
# cmd session JSONL -> canonical events
cmd session events -> FromCmd -> []types.AgentEvent

# canonical events -> cmd session events
[]types.AgentEvent -> ToCmd -> cmd session events
```

## Preconditions
- `FromCmd` and `ToCmd` functions are defined in `agent/event/cmd_types`.
- Input is provided as JSONL in `CmdInput` or as `[]types.AgentEvent` in `Events`.

## Steps
1. Leaf SETUPs populate `req.CmdInput` or `req.Events`.
2. Group SETUPs set `req.Target` to `"from_cmd"` or `"cmd"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}
```
