# Claude ClaudeEventWriter Tests

Unit tests for `ClaudeEventWriter` in `agent/cli/claude/claude.go`, which
converts each `claude` headless NDJSON line into canonical `AgentEvent` JSONL
via the `agent/event/claude_types` package. No `claude` CLI binary required.

## Version

0.0.2

## DSN (Domain Specific Notion)

Participants:

- **ClaudeEventWriter** — wraps a `RawLog` writer. Each call to
  `WriteClaudeLine(line)` parses one native NDJSON line and emits zero or more
  canonical `AgentEvent` JSONL lines. `Flush()` finalizes any buffered state.
- **claude_types.StreamEvent** — the native line model
  (`system`/`assistant`/`user`/`result`).
- **claude_types.FromClaude** — the mapping: `system init` → `step_start`;
  `assistant` text/thinking/tool_use blocks → `message`/`think`/`tool_call`
  (in array order); `user` tool_result → skipped; `result` success → `done`,
  `result` error → `error`.
- **AgentEvent** — the canonical event written to `RawLog`, one JSON object
  per line.

Behaviors:

- A single `assistant` line with multiple content blocks emits one `AgentEvent`
  per block, in array order.
- A `user` tool_result line emits nothing.
- A `result` line emits exactly one terminal event (`done` or `error`).

## Decision Tree

```
write-events/
├── system-init/              system init line → 1 step_start
├── assistant-mixed-blocks/   1 assistant line (text+thinking+tool_use) → 3 events
│                             (message, think, tool_call) in order
├── user-tool-result/         user tool_result line → 0 events (empty)
├── result-success/           result success line → 1 done
└── result-error/             result error line → 1 error
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `system-init` | A `system` init NDJSON line → exactly 1 `step_start` AgentEvent |
| 2 | `assistant-mixed-blocks` | One `assistant` line with text+thinking+tool_use blocks → exactly 3 AgentEvents (`message`, `think`, `tool_call`) in order |
| 3 | `user-tool-result` | A `user` tool_result NDJSON line → zero AgentEvents (lines empty) |
| 4 | `result-success` | A `result` success line → exactly 1 `done` AgentEvent |
| 5 | `result-error` | A `result` error line → exactly 1 `error` AgentEvent |

## How to Run

```sh
doctest test ./agent/cli/claude/tests/write-events/...
```

```go
import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/claude"
	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

type Request struct {
	ClaudeLines []string
}

type Response struct {
	Lines []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var buf bytes.Buffer
	w := claude.NewClaudeEventWriter(&buf)
	for _, line := range req.ClaudeLines {
		w.WriteClaudeLine(line)
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
