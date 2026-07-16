# Scenario

**Feature**: send follow-up message to live commandcode-tty session

```
agent-run run --open --session-id <id> "Hello"   (background)
agent-run send --max-wait 15s <id> "Hello 2"
agent-run tty snapshot <id>  → contains both prompts
```

## Steps

1. Start `run --open` with "Hello" in the background.
2. Send "Hello 2" with `--max-wait 15s`.
3. Take snapshot and verify both prompts appear.

```go
import (
	"fmt"
	"os/exec"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = fmt.Sprintf("e2e-%d", time.Now().UnixNano())
	return nil
}
```
