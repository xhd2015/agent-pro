# Scenario

**Feature**: writeAgentEventsFromGrokLine converts grok streaming JSON lines to AgentEvent JSONL

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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = req.GrokLines
	return nil
}
```
