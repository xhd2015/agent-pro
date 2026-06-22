## Preconditions
- `writeAgentEventsFromGrokLine` converts each grok streaming JSON line to AgentEvent JSONL.
- Grok `thought` events arrive word-by-word; each line should eventually coalesce to one think event.

## Steps
1. Set `req.GrokLines` to grok native streaming JSON lines.
2. Feed each line through one shared `grok.NewGrokEventWriter`, then `Flush()`.
3. Return marshaled AgentEvent lines in `resp.Lines`.

```go
import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/grok"
	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	GrokLines []string
}

type Response struct {
	Lines []string
}

func Setup(t *testing.T, req *Request) error {
	_ = req.GrokLines
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var buf bytes.Buffer
	w := grok.NewGrokEventWriter(&buf)
	for _, line := range req.GrokLines {
		w.WriteGrokLine(line)
	}
	w.Flush()
	var lines []string
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var ev eventtypes.AgentEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			t.Fatalf("unmarshal agent event: %v", err)
		}
		lines = append(lines, string(line))
	}
	return &Response{Lines: lines}, nil
}
```