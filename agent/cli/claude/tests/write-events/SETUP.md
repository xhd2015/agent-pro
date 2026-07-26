# Scenario

**Feature**: ClaudeEventWriter converts each native claude NDJSON line into canonical AgentEvent JSONL

```
# feed native claude stream-json lines, collect AgentEvent JSONL on a buffer
write-events -> ClaudeEventWriter.WriteClaudeLine(line) -> RawLog (AgentEvent JSONL)
write-events -> ClaudeEventWriter.Flush() -> RawLog (finalized)
Run <- RawLog (Lines []string of AgentEvent JSON)
```

## Preconditions
- `ClaudeEventWriter` converts each claude streaming JSON line to AgentEvent
  JSONL via `claude_types.FromClaude`.
- A `user` tool_result line emits no AgentEvent.

## Steps
1. Set `req.ClaudeLines` to claude native streaming JSON lines.
2. Feed each line through one shared `claude.NewClaudeEventWriter`, then
   `Flush()`.
3. Return marshaled AgentEvent lines in `resp.Lines`.

## Context
- No `claude` binary is required; this is a pure unit test of the writer.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/claude"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = req.ClaudeLines
	_ = (*claude.ClaudeAgent)(nil)
	return nil
}
```
