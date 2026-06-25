## Preconditions
- The `pi` binary is available in PATH.
- PI must have API keys configured in its own config (`~/.pi/agent/`).
- This runs real queries against the pi CLI through PiAgent.

## Steps
1. Look up the pi binary from PATH using `pi.FindAgentPath`; skip if not installed.
2. Initialize the PiAgent with the resolved binary path.
3. Execute the query via `agent.Ask()` and return the answer.

```go
import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/pi"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

func Setup(t *testing.T, req *Request) error {
	req.Model = os.Getenv("PI_MODEL")
	return nil
}

func parseRawLog(buf bytes.Buffer) []json.RawMessage {
	var events []json.RawMessage
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		line = line[:len(line)-1] // trim newline
		if line == "" {
			continue
		}
		events = append(events, json.RawMessage(line))
	}
	return events
}
```
