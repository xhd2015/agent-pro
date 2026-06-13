## Preconditions
- The program imports `agent/event/types` and prints all exported `ActionType` constants.

## Steps
1. Run a program that prints each constant name and value.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"fmt"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func main() {
	fmt.Printf("ActionThink=%s\n", types.ActionThink)
	fmt.Printf("ActionToolCall=%s\n", types.ActionToolCall)
	fmt.Printf("ActionMessage=%s\n", types.ActionMessage)
	fmt.Printf("ActionError=%s\n", types.ActionError)
	fmt.Printf("ActionDone=%s\n", types.ActionDone)
}
`
	return nil
}
```
