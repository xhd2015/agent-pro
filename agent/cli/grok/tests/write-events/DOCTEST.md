# Grok writeAgentEventsFromGrokLine Tests

Unit tests for `writeAgentEventsFromGrokLine` in `agent/cli/grok/grok.go`.
No grok CLI binary required.

## Decision Tree

```
write-events/
├── DOCTEST.md
├── SETUP.md
├── thought-streaming-deltas/   Per-word thought lines → 1 coalesced think event (RED)
└── tool-call-streaming-lines/  tool_started/completed lines → tool_call AgentEvents (RED)
```

## How to Run

```sh
doctest test ./agent/cli/grok/tests/write-events/...
```

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
