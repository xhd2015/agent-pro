# Session activity

## Version

0.0.1

# DSN (Domain Specific Notion)

**Participants**

- **Durable log** — DSH v3 events in order (`turn/start`, `assistant/attempt`,
  `assistant/message`, `tool/call`, `tool/result`, `turn/end`, `user/message`).
- **Last chat message** — grok/codex coalesced kind (`thinking` | `tool` |
  `response` | `user`) plus optional tool name.
- **Liveness** — whether a write lease / runner PID is held.
- **Activity** — `idle` | `thinking` | `responding` | `tool_calling` | `error`.

**Behaviors**

- DSH open turn with no assistant text → thinking.
- DSH assistant text before `turn/end` → responding.
- DSH `tool/call` without `tool/result` → tool_calling (detail = tool name).
- DSH `turn/end` completed → idle.
- DSH open turn and not live → error.
- Chat last kind thinking + live → thinking; thinking + not live → error.
- Chat last kind tool + live → tool_calling.
- Chat last kind response → idle (TUI may still be open).

## How to Run

```sh
doctest test ./tests/session-activity
```

```go
import (
	"encoding/json"
	"testing"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
	"github.com/xhd2015/agent-pro/agent/sessionactivity"
	"github.com/xhd2015/doctest/session"
)

type EventSpec struct {
	Type string
	Data string
}

type Request struct {
	Source   string
	Events   []EventSpec
	Live     bool
	LastKind string
	Tool     string
}

type Response struct {
	Kind   string
	Detail string
	Live   bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.Source == "chat" {
		kind, detail := sessionactivity.ClassifyChat(req.LastKind, req.Tool, req.Live)
		return &Response{Kind: string(kind), Detail: detail, Live: req.Live}, nil
	}
	events := make([]dshsessions.Event, 0, len(req.Events))
	for _, spec := range req.Events {
		events = append(events, dshsessions.Event{Type: spec.Type, Data: json.RawMessage(spec.Data)})
	}
	act := dshsessions.Classify(events, req.Live)
	return &Response{Kind: string(act.Kind), Detail: act.Detail, Live: act.Live}, nil
}
```
