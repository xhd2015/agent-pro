# Scenario

**Feature**: The program imports `agent/event/types` and prints all exported `ActionType` constants

## Preconditions
- The program imports `agent/event/types` and prints all exported `ActionType` constants.

## Steps
1. Run a program that prints each constant name and value.

```go
import (
	"fmt"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "ActionThink=%s\n", types.ActionThink)
	fmt.Fprintf(&sb, "ActionToolCall=%s\n", types.ActionToolCall)
	fmt.Fprintf(&sb, "ActionMessage=%s\n", types.ActionMessage)
	fmt.Fprintf(&sb, "ActionError=%s\n", types.ActionError)
	fmt.Fprintf(&sb, "ActionDone=%s\n", types.ActionDone)
	req.Output = sb.String()
	return nil
}
```
