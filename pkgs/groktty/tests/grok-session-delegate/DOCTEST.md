# groktty grok_session delegate tests

Doc-style tests verifying that `pkgs/groktty` delegates ACP `updates.jsonl`
conversion to `agent/event/grok_session`. Public APIs under test:

- `TailUpdates` / `TailUpdatesFromOffset` — poll on-disk updates and emit
  canonical `types.AgentEvent` with grok_session fields
- `DiscoverSession` — locate grok session dirs by cwd + first user prompt chunk
  (uses `grok_session.ParseLine` after refactor)

No real grok binary, no PTY. Synthetic `updates.jsonl` fixtures only.

## Version

0.0.2

# DSN (Domain Specific Notion)

Grok persists conversations under `$GROK_HOME/sessions/<encoded-cwd>/<session-uuid>/`.
Each session dir has `summary.json` (cwd, created_at) and `updates.jsonl` (ACP wire).

Participants and behaviors:

- **updates.jsonl** — append-only JSONL of ACP session updates (flat or
  `_x.ai/session/update` envelope).
- **grok_session.Converter** — stateful forward adapter; coalesces chunks, stamps
  `tool_call_id`, `extensions.grok_session.status`, `extensions.grok_session.turn_index`,
  emits `ActionDone` at `turn_completed`.
- **TailUpdatesFromOffset** — polls `updates.jsonl` from a byte offset, feeds each
  new line through the converter, emits events via callback; stops after
  `turn_completed` or context cancel.
- **DiscoverSession** — scans session dirs matching workspace cwd and
  `created_at >= runStart`; matches prompt against the first `user_message_chunk`
  text extracted via `grok_session.ParseLine`.
- **Integration check** — events collected from tail must semantically match
  `grok_session.FromUpdatesJSONL` on the same wire fixture.

## Decision Tree

```
Level 1: API surface
├── tail/           TailUpdatesFromOffset emits grok_session-rich events
├── session/        DiscoverSession prompt matching via grok_session.ParseLine
└── integration/    tail output ≡ grok_session.FromUpdatesJSONL semantics

Level 2 (tail): grok_session field coverage
├── emits-tool-call-id          tool_call + tool_call_update → matching tool_call_id
├── emits-grok-session-status   status pending then completed on tool events
├── emits-turn-index-and-done   turn_index=0 on all events; ActionDone at turn end
└── nested-envelope-line        envelope wire user/assistant text emitted

Level 2 (session): discovery path
└── prompt-matches-first-user-chunk   nested envelope first user chunk matches prompt

Level 2 (integration): cross-package parity
└── events-jsonl-rich-fields    tail events SemanticEqual grok_session.FromUpdatesJSONL
```

Parameter ranking (most → least significant):

1. **API** — tail vs session discovery vs integration parity
2. **Wire shape** — flat ACP vs nested envelope
3. **Event richness** — tool_call_id / status / turn_index / ActionDone
4. **Fixture scope** — tool-only vs full turn sequence

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `tail/emits-tool-call-id` | `TailUpdatesFromOffset` emits tool events with matching `tool_call_id` |
| 2 | `tail/emits-grok-session-status` | Tool events carry `extensions.grok_session.status` pending then completed |
| 3 | `tail/emits-turn-index-and-done` | Full turn: all events `turn_index=0`; `ActionDone` at `turn_completed` |
| 4 | `tail/nested-envelope-line` | Envelope wire line tails; user and assistant text emitted |
| 5 | `session/prompt-matches-first-user-chunk` | `DiscoverSession` matches prompt via nested envelope user chunk |
| 6 | `integration/events-jsonl-rich-fields` | Tail-collected events semantically equal `grok_session.FromUpdatesJSONL` |

## How to Run

```sh
doctest vet ./pkgs/groktty/tests/grok-session-delegate
doctest test ./pkgs/groktty/tests/grok-session-delegate
doctest test -v ./pkgs/groktty/tests/grok-session-delegate/tail/emits-tool-call-id
doctest test ./pkgs/groktty/tests/grok-session-delegate/integration/...
```

```go
import (
	"context"
	"testing"
	"time"

	grok_session "github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

type Request struct {
	Target      string // "tail", "session", "integration"
	WireLines   []string
	UpdatesPath string
	GrokHome    string
	Workspace   string
	SessionUUID string
	SessionID   string // envelope session id
	Prompt      string
	RunStart    time.Time
}

type Response struct {
	Events         []types.AgentEvent
	ExpectedEvents []types.AgentEvent
	SessionID      string
	UpdatesPath    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Target {
	case "tail", "integration":
		if req.UpdatesPath == "" {
			t.Fatal("tail/integration requires UpdatesPath from Setup")
		}
		if err := writeUpdatesJSONL(req.UpdatesPath, req.WireLines...); err != nil {
			return nil, err
		}
		events, err := tailAllEvents(t, req.UpdatesPath)
		if err != nil {
			return nil, err
		}
		resp := &Response{Events: events, UpdatesPath: req.UpdatesPath}
		if req.Target == "integration" {
			resp.ExpectedEvents = grok_session.FromUpdatesJSONL(req.WireLines)
		}
		return resp, nil
	case "session":
		if req.GrokHome == "" || req.Workspace == "" || req.Prompt == "" {
			t.Fatal("session requires GrokHome, Workspace, and Prompt from Setup")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		id, path, err := groktty.DiscoverSession(ctx, req.GrokHome, req.Workspace, req.Prompt, req.RunStart)
		if err != nil {
			return nil, err
		}
		return &Response{SessionID: id, UpdatesPath: path}, nil
	default:
		t.Fatalf("unknown Target %q", req.Target)
	}
	return nil, nil
}
```