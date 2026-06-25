# FormatState Streaming Tests

These doc-style tests verify `print.FormatState.FormatLine` streaming coalescing
for consecutive `AgentEvent` deltas. Used by `traceSession` when rendering
events.jsonl lines.

## Decision Tree

```
format-state-streaming/
├── DOCTEST.md
├── SETUP.md
├── grok-thought-deltas/        ActionThink per-word deltas → 1 think block (RED)
└── message-deltas-coalesced/   ActionMessage deltas → 1 ASSISTANT block
```

## How to Run

```sh
doctest test ./agent/event/print/tests/format-state-streaming/...
```

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/event/print"
)


type Request struct {
	Lines []string
}

type Response struct {
	Output string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var state print.FormatState
	var buf strings.Builder
	for _, line := range req.Lines {
		header, body, isMsg := state.FormatLine(line)
		if header == "" && body == "" && !isMsg {
			continue
		}
		if isMsg {
			if header != "" {
				buf.WriteString(header)
				buf.WriteByte('\n')
			}
			buf.WriteString(body)
		} else {
			if header != "" {
				buf.WriteString(header)
				buf.WriteByte('\n')
			}
			if body != "" {
				buf.WriteString(body)
			}
		}
	}
	state.Flush()
	return &Response{Output: buf.String()}, nil
}
```
