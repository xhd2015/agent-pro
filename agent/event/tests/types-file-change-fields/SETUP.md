## Preconditions
- The program imports `agent/event/types` and marshals a `FileChange`.

## Steps
1. Create a FileChange with path and kind.
2. Marshal to JSON and print.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func main() {
	fc := types.FileChange{Path: "bar.go", Kind: "modify"}
	data, _ := json.Marshal(fc)
	fmt.Println(string(data))
}
`
	return nil
}
```
