# Grok writeAgentEventsFromGrokLine Tests

Unit tests for `writeAgentEventsFromGrokLine` in `agent/cli/grok/grok.go`.
No grok CLI binary required.

## Version

0.0.2

## DSN (Domain Specific Notion)

Participants:

- **writeAgentEventsFromGrokLine** — the pure conversion function that maps one
  native grok streaming-JSON line to zero or more canonical `AgentEvent` JSONL
  lines. It is the unit under test; no binary is spawned.
- **GrokEventWriter** — the shared writer (`grok.NewGrokEventWriter`) that
  buffers incoming native lines and flushes them as `AgentEvent`s. Each test
  feeds a sequence of `GrokLines` through one writer, then `Flush()`.
- **grok native lines** — JSON objects emitted by the grok CLI stream:
  `thought` (per-word reasoning deltas that must coalesce), `text` (assistant
  content), and `tool_started`/`tool_completed` (tool activity currently
  dropped by the writer).

Behaviors:

- Per-word `thought` deltas accumulate and flush to a single `ActionThink`
  AgentEvent.
- `text` lines flush to `ActionMessage` AgentEvents.
- `tool_started`/`tool_completed` lines are dropped today (RED leaves track the
  desired `ActionToolCall` conversion).
- Output is one JSON object per line on the writer's buffer, parsed back as
  `AgentEvent` for assertion.

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
	"github.com/xhd2015/doctest/session"
)


type Request struct {
	GrokLines []string
}

type Response struct {
	Lines []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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
