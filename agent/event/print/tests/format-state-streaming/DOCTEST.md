# FormatState Streaming Tests

# DSN (Domain Specific Notion)

`FormatState.FormatLine` renders one JSONL `AgentEvent` at a time while coalescing
consecutive assistant message and think deltas into a single streaming block.
Tool-call events flush any open message/think block and emit a standalone formatted
header. Used by `traceSession` when walking `events.jsonl`.

## Version

0.0.2

## Decision Tree

```
format-state-streaming/
├── DOCTEST.md
├── SETUP.md
├── grok-thought-deltas/              ActionThink per-word deltas → 1 think block (RED)
├── message-deltas-coalesced/         ActionMessage deltas → 1 ASSISTANT block
├── skill-tool-call-details/          skill tool_call events → SKILL blocks with names
└── maintain-topic-web-todo-details/    web search, todo, webfetch → blocks with input details
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