# Event Package Tests

These doc-style tests verify the `agent/event/types`, `agent/event/codex_types`, `agent/event/opencode_types`, `agent/event/crush_types`, and `agent/event/claude_types` packages.

## Version

0.0.2

# DSN (Domain Specific Notion)

The event layer models a CLI agent run as a stream of **canonical `AgentEvent`s**
(the common currency of `agent/event/types`). Each per-CLI `*_types` package is an
adapter between one vendor's native event format and that canonical stream.

Participants and behaviors:

- **Canonical AgentEvent** — the shared representation. Each event has a `type`
  drawn from a fixed set: `step_start`, `think`, `message`, `tool_call`,
  `error`, `done` (plus phase/role/text/tool fields). All leaves ultimately
  assert on JSON-marshaled canonical events.
- **Vendor native stream** — the wire format a CLI emits. The adapter knows how
  to walk it: `codex_types` (item.* events), `opencode_types`, `crush_types`
  (SSE `{type,payload}` envelopes), `claude_types` (headless NDJSON with
  `system`/`assistant`/`user`/`result` lines), and `cmd_types` (JSONL session
  events with `user`/`assistant`/`tool` roles and typed content blocks).
- **From<Vendor>** — native → canonical. Walks the native stream and emits zero
  or more `AgentEvent`s per native event (some native constructs map to nothing,
  e.g. tool-result echoes that are folded back into the preceding tool call).
- **To<Vendor>** — canonical → native. The inverse: each `AgentEvent` becomes
  one native event of the appropriate shape.
- **Integration leaves** — start the real CLI binary (or model it headless),
  feed a prompt, capture the native stream, run it through `From<Vendor>`, and
  assert the canonical output contains the expected answer. Skipped when the
  binary is absent or `<VENDOR>_SKIP_INTEGRATION=1`.

The `Run` function in the Go block below is the single dispatch point: it reads
`req.Target` and the populated `Request` fields to decide which adapter or
integration helper to invoke, then marshals the canonical result to
`resp.Output`.

## Decision Tree: Crush Event Types

The crush runner uses SSE payloads with `{\"type\": \"<payload_type>\", \"payload\": {...}}`.
Conversion functions `FromCrush` (crush→canonical) and `ToCrush` (canonical→crush) are tested.

```
crush_types/
├── to_crush/                        (Target="crush", canonical→crush)
│   ├── think/                       ActionThink → message with reasoning part
│   ├── message/                     ActionMessage → message with text part
│   ├── tool_call/                   ActionToolCall → message with tool_call part
│   ├── error/                       ActionError → agent_event with type=error
│   └── done/                        ActionDone → run_complete
│
└── from_crush/                      (Target="from_crush", crush→canonical)
    ├── reasoning/                    reasoning part → ActionThink
    ├── text/                         text part → ActionMessage
    ├── tool_call/                    tool_call part → ActionToolCall
    ├── tool_result/                  tool_result part → skipped (no action)
    ├── finish_error/                 finish reason=error → ActionError
    ├── finish_no_error/              finish reason=end_turn → skipped
    ├── mixed_parts/                  text + tool_call → multiple actions
    ├── non_assistant/                user/system role → skipped
    ├── agent_event_error/            agent_event type=error → ActionError
    ├── agent_event_response/         agent_event type=response → no ActionError
    ├── run_complete/                 run_complete success → ActionDone
    ├── run_complete_error/           run_complete with error → ActionDone
    └── run_complete_cancelled/       run_complete cancelled → ActionDone
```

## Test Leaves

| Leaf | Description |
|---|---|
| `crush_types-to-crush-think` | ToCrush: ActionThink → crush reasoning message |
| `crush_types-to-crush-message` | ToCrush: ActionMessage → crush text message |
| `crush_types-to-crush-tool-call` | ToCrush: ActionToolCall → crush tool_call message |
| `crush_types-to-crush-error` | ToCrush: ActionError → crush agent_event error |
| `crush_types-to-crush-done` | ToCrush: ActionDone → crush run_complete |
| `crush_types-from-crush-reasoning` | FromCrush: reasoning part → ActionThink |
| `crush_types-from-crush-text` | FromCrush: text part → ActionMessage |
| `crush_types-from-crush-tool-call` | FromCrush: tool_call part → ActionToolCall |
| `crush_types-from-crush-tool-result` | FromCrush: tool_result part → skipped |
| `crush_types-from-crush-finish-error` | FromCrush: finish error → ActionError |
| `crush_types-from-crush-finish-no-error` | FromCrush: finish end_turn → no action |
| `crush_types-from-crush-mixed-parts` | FromCrush: text + tool_call in one message |
| `crush_types-from-crush-non-assistant` | FromCrush: user/system message → skipped |
| `crush_types-from-crush-agent-event-error` | FromCrush: agent_event error → ActionError |
| `crush_types-from-crush-agent-event-response` | FromCrush: agent_event response → no error |
| `crush_types-from-crush-run-complete` | FromCrush: run_complete success → ActionDone |
| `crush_types-from-crush-run-complete-error` | FromCrush: run_complete with error text |
| `crush_types-from-crush-run-complete-cancelled` | FromCrush: run_complete cancelled=true |

## Decision Tree: Crush Server-Mode SSE Integration

The `Target="crush_server"` mode starts a live crush server, sends prompts
via HTTP, captures SSE events, and runs them through `FromCrush`.

```
crush_integration/
├── SETUP.md (grouping: Target="crush_server", ModelName="deepseek-v4-pro")
│
├── crush-integration-real/            ✓ Happy path: start server, prompt,
│   │                                       capture SSE, verify "paris"
│   ├── SETUP.md: Prompt="one word of French capital", HostPort=0
│   └── ASSERT.md: Output contains ActionMessage with "paris" (case-insensitive)
│
└── crush-integration-binary-not-found/  ✗ Explicit CrushPath not on disk
    ├── SETUP.md: CrushPath="/nonexistent/path/to/crush"
    └── ASSERT.md: non-nil error referencing missing binary
```

### New Test Leaves

| Leaf | Description |
|---|---|
| `crush-integration-real` | End-to-end: start crush server, prompt, SSE capture, FromCrush, verify "paris" |
| `crush-integration-binary-not-found` | Explicit CrushPath points to nonexistent file → error or skip |

## Decision Tree: Claude Event Types

The Claude Code **headless** runner (`claude -p --output-format stream-json --verbose`)
emits one JSON object per stdout line with top-level `type` of `system`,
`assistant`, `user`, or `result`. Conversion functions `FromClaude`
(claude→canonical) and `ToClaude` (canonical→claude) are tested.

```
claude_types/
├── claude_types-event-types/        leaf — constants + StreamEvent marshaling
│
├── from_claude/                     (Target="from_claude", claude→canonical)
│   ├── system-init/                 system init → ActionStepStart
│   ├── assistant-text/              assistant text block → ActionMessage
│   ├── assistant-thinking/          thinking block → ActionThink
│   ├── assistant-tool-use/          tool_use block → ActionToolCall
│   ├── user-tool-result/            user tool_result → skipped (no action)
│   ├── result-success/              result success → ActionDone
│   ├── result-error/                result error → ActionError
│   └── mixed/                       text + thinking + tool_use in one assistant msg
│
├── to_claude/                       (Target="claude", canonical→claude)
│   ├── think/                       ActionThink → assistant thinking event
│   ├── message/                     ActionMessage → assistant text event
│   ├── tool-call/                   ActionToolCall → assistant tool_use event
│   ├── error/                       ActionError → result error event
│   └── done/                        ActionDone → result success event
│
└── claude_integration/              (Target="claude_headless")
    ├── claude-integration-real/            ✓ live: claude -p "say pong", parse, verify "pong"  [slow && heavy]
    └── claude-integration-binary-not-found/  ✗ ClaudePath="/nonexistent/claude" → error
```

### Claude Test Leaves

| Leaf | Description |
|---|---|
| `claude_types-event-types` | Constants + StreamEvent marshaling round-trip |
| `claude_types-from-claude-system-init` | FromClaude: system init → ActionStepStart |
| `claude_types-from-claude-assistant-text` | FromClaude: assistant text block → ActionMessage |
| `claude_types-from-claude-assistant-thinking` | FromClaude: thinking block → ActionThink |
| `claude_types-from-claude-assistant-tool-use` | FromClaude: tool_use block → ActionToolCall |
| `claude_types-from-claude-user-tool-result` | FromClaude: user tool_result → skipped (no action) |
| `claude_types-from-claude-result-success` | FromClaude: result success → ActionDone |
| `claude_types-from-claude-result-error` | FromClaude: result error → ActionError |
| `claude_types-from-claude-mixed` | FromClaude: text + thinking + tool_use in one assistant message |
| `claude_types-to-claude-think` | ToClaude: ActionThink → assistant thinking event |
| `claude_types-to-claude-message` | ToClaude: ActionMessage → assistant text event |
| `claude_types-to-claude-tool-call` | ToClaude: ActionToolCall → assistant tool_use event |
| `claude_types-to-claude-error` | ToClaude: ActionError → result error event |
| `claude_types-to-claude-done` | ToClaude: ActionDone → result success event |
| `claude-integration-real` | End-to-end: claude -p "say pong", stream-json, FromClaude, verify "pong" |
| `claude-integration-binary-not-found` | Explicit ClaudePath points to nonexistent file → non-nil error |

### Decision Tree: Cmd Event Types

The `cmd` CLI (Command Code) stores sessions as JSONL files. Each line has `role` (`user`/`assistant`/`tool`)
and `content` (an array of typed blocks). Conversion functions `FromCmd` (cmd→canonical) and `ToCmd`
(canonical→cmd) are tested.

```
cmd_types/
├── from_cmd/                        (Target="from_cmd", cmd→canonical)
│   ├── assistant-text/              assistant text block → ActionMessage
│   ├── assistant-reasoning/         assistant reasoning block → ActionThink
│   ├── assistant-tool-call/         assistant tool-call block → ActionToolCall
│   ├── reasoning-text-tool-call/    reasoning + text + tool-call → 3 events
│   ├── tool-call-with-result/       tool-call + tool-result → ActionToolCall with Output
│   └── user-text/                   user text → ActionStepStart
│
└── to_cmd/                          (Target="cmd", canonical→cmd)
    ├── think/                        ActionThink → assistant reasoning block
    ├── message/                      ActionMessage → assistant text block
    └── tool-call/                    ActionToolCall → assistant tool-call block
```

### Cmd Test Leaves

| Leaf | Description |
|---|---|
| `cmd_types-from-cmd-assistant-text` | FromCmd: assistant text → ActionMessage |
| `cmd_types-from-cmd-assistant-reasoning` | FromCmd: assistant reasoning → ActionThink |
| `cmd_types-from-cmd-assistant-tool-call` | FromCmd: assistant tool-call → ActionToolCall |
| `cmd_types-from-cmd-reasoning-text-tool-call` | FromCmd: reasoning + text + tool-call → 3 canonical events |
| `cmd_types-from-cmd-tool-call-with-result` | FromCmd: tool-call then tool-result → merged output |
| `cmd_types-from-cmd-user-text` | FromCmd: user text → ActionStepStart |
| `cmd_types-to-cmd-think` | ToCmd: ActionThink → assistant reasoning block |
| `cmd_types-to-cmd-message` | ToCmd: ActionMessage → assistant text block |
| `cmd_types-to-cmd-tool-call` | ToCmd: ActionToolCall → assistant tool-call block |

## How to Run

```sh
doctest test ./agent/event/tests
```

```go
import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	cmd_types "github.com/xhd2015/agent-pro/agent/event/cmd_types"
	codex_types "github.com/xhd2015/agent-pro/agent/event/codex_types"
	claude_types "github.com/xhd2015/agent-pro/agent/event/claude_types"
	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)


type Request struct {
	Events      []types.AgentEvent
	Target      string // "opencode"→ToOpencode, "crush"→ToCrush, "from_crush"→FromCrush, "crush_server"→StartCrushServer, "claude"→ToClaude, "from_claude"→FromClaude, "claude_headless"→runClaudeHeadless, "cmd"→ToCmd, "from_cmd"→FromCmd; default→ToCodex
	SessionID   string
	Value       any
	Output      string
	CrushInput  string // raw JSON for FromCrush parsing
	ClaudeInput string // raw NDJSON (one JSON object per line) for FromClaude parsing
	CmdInput    string // raw JSON (one JSON object per line) for FromCmd parsing
	HostPort    int    // crush server HTTP port (0 = auto-assign)
	Prompt      string // prompt text for crush server / claude headless
	ModelName   string // model override (default: "deepseek-v4-pro")
	CrushPath   string // path to crush binary (default: LookPath("crush"))
	ClaudePath  string // path to claude binary (default: LookPath("claude"))
}

type Response struct {
	Output string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var output string
	if req.Target == "from_crush" && req.CrushInput != "" {
		var crushEvents []crush_types.Event
		if err := json.Unmarshal([]byte(req.CrushInput), &crushEvents); err != nil {
			return &Response{Output: ""}, nil
		}
		result := crush_types.FromCrush(crushEvents, req.SessionID)
		data, _ := json.Marshal(result)
		output = string(data)
	} else if req.Target == "from_claude" && req.ClaudeInput != "" {
		var claudeEvents []claude_types.StreamEvent
		for _, line := range strings.Split(req.ClaudeInput, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var ev claude_types.StreamEvent
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				return &Response{Output: ""}, nil
			}
			claudeEvents = append(claudeEvents, ev)
		}
		result := claude_types.FromClaude(claudeEvents, req.SessionID)
		data, _ := json.Marshal(result)
		output = string(data)
	} else if req.Target == "from_cmd" && req.CmdInput != "" {
		var cmdEvents []cmd_types.Event
		for _, line := range strings.Split(req.CmdInput, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var ev cmd_types.Event
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				return &Response{Output: ""}, nil
			}
			cmdEvents = append(cmdEvents, ev)
		}
		result := cmd_types.FromCmd(cmdEvents, req.SessionID)
		data, _ := json.Marshal(result)
		output = string(data)
	} else if req.Target == "cmd" && len(req.Events) > 0 {
		result := cmd_types.ToCmd(req.Events, req.SessionID)
		data, _ := json.Marshal(result)
		output = string(data)
	} else if req.Target == "claude" && len(req.Events) > 0 {
		result := claude_types.ToClaude(req.Events, req.SessionID)
		data, _ := json.Marshal(result)
		output = string(data)
	} else if len(req.Events) > 0 {
		if req.Target == "opencode" {
			result := opencode_types.ToOpencode(req.Events, req.SessionID)
			data, _ := json.Marshal(result)
			output = string(data)
		} else if req.Target == "crush" {
			result := crush_types.ToCrush(req.Events, req.SessionID)
			data, _ := json.Marshal(result)
			output = string(data)
		} else {
			result := codex_types.ToCodex(req.Events)
			data, _ := json.Marshal(result)
			output = string(data)
		}
	} else if req.Value != nil {
		data, _ := json.Marshal(req.Value)
		output = string(data)
	} else if req.Target == "crush_server" {
		return runCrushServer(t, req)
	} else if req.Target == "claude_headless" {
		return runClaudeHeadless(t, req)
	} else {
		output = req.Output
	}
	return &Response{Output: output}, nil
}
```
