# Print Package Tests

These doc-style tests verify the `agent/event/print` package. The package
parses agent event JSONL lines and formats them as compact human-readable
strings.

Tests import the print package and the opencode adapter (for side-effect
adapter registration) and call `FormatTraceLine` with various JSONL event
lines.

## Decision Tree

The `format-state-streaming/` subtree tests `FormatState.FormatLine` streaming
coalescing for consecutive AgentEvent deltas (grok thought, message streaming).

The `pi_types/` subtree tests the Pi trace adapter, which handles Pi agent
events (type:\`message_start\`, \`tool_execution_start\`, etc.).

```
format-state-streaming/                  FormatState streaming coalescing
├── grok-thought-deltas                ActionThink per-word deltas → 1 think block (RED)
└── message-deltas-coalesced           ActionMessage deltas → 1 ASSISTANT block

pi_types/                              Root: Pi adapter registration
├── session                            type=session → skip
├── non-assistant                      message_update, role=user → skip
├── message-start                      message_start, role=assistant, text → ASSISTANT
├── message-update-text               message_update, role=assistant, text_delta → ASSISTANT
├── message-update-thinking           message_update, role=assistant, thinking_delta → ASSISTANT
├── message-end                        message_end, role=assistant, text → ASSISTANT
├── tool-exec-start                    tool_execution_start, bash → RUN
├── tool-exec-end-ok                   tool_execution_end, ok → RUN + result
└── tool-exec-end-err                  tool_execution_end, error → RUN + FAILED
```

### Leaf Index

| Leaf | Event Type | Expected Output |
|---|---|---|
| `session` | `session` | empty |
| `non-assistant` | `message_update` with `role:user` | empty |
| `message-start` | `message_start` assistant text | `ASSISTANT` + `Hello` |
| `message-update-text` | `message_update` text_delta, delta = `"world"` | `ASSISTANT` + `world` (uses Delta) |
| `message-update-accumulated-text` | `message_update` with large Content, small delta | `ASSISTANT` + ` feature.` (delta only, not accumulated) |
| `message-update-thinking` | `message_update` thinking_delta, delta = `"hmm"` | `ASSISTANT` + `hmm` (uses Delta) |
| `message-end` | `message_end` assistant text, no delta | empty (deltas already shown) |
| `tool-exec-start` | `tool_execution_start` bash | `RUN` + `ls -la` |
| `tool-exec-end-ok` | `tool_execution_end` ok | `RUN` + `file1.txt` |
| `tool-exec-end-err` | `tool_execution_end` error | `RUN` + `not found` + `FAILED` |

## How to Run

```sh
# All print package tests (existing + pi)
doctest test ./agent/event/print/tests/...

# Pi adapter tests only
doctest test ./agent/event/print/tests/pi_types
```

```go
import (

	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/event/print"
	"github.com/xhd2015/doctest/session"
)


type Request struct {
	Line string
}

type Response struct {
	Output string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return &Response{Output: print.FormatTraceLine(req.Line)}, nil
}
```
