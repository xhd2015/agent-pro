# Trace Parse Consolidation Tests

Doc-style tests for moving `agent_trace` parsing into `agent/event/traceparse`,
`agent/event/traceview`, and `agent/event/summary`, while keeping `agent_trace` as a
thin backward-compatible wrapper.

# DSN (Domain Specific Notion)

**Participants**

- **Trace JSONL line** — one provider-native JSON object from `events.jsonl`.
- **Adapter registry** — ordered list of provider parsers (codex, cursor, opencode, pi, generic).
- **Line parser** — picks the first adapter that claims a line; yields `AgentTraceParsedEvent`
  (assistant message and/or tool activity).
- **Message aggregator** — walks many lines, merges tool activities by `call_id` / tool name,
  assigns synthetic timestamps from trace `created_at`.
- **Summary helpers** — compact long tool output and humanize tool identifiers.
- **Print formatter** — `FormatTraceLine` tries canonical `AgentEvent` first, then adapter fallback.
- **Thin wrapper** — `agent_trace` re-exports types and `ParseMessages`, blank-imports adapters.

**Behaviors**

- A single line enters the registry; the winning adapter returns message or activity semantics.
- Multiple lines enter the aggregator; started/updated/completed subtypes merge into timeline messages.
- `agent_trace` public API remains stable via aliases and delegation.
- Print path must not depend on `agent_trace` adapter packages after consolidation.

## Version

0.0.2

## Decision Tree

```
Level 1: Mode (req.Mode)
├── parse-line/              one JSONL line → AgentTraceParsedEvent (traceparse)
│   ├── codex/               item.* Codex rollout items
│   │   ├── plan-todo        todo_list updated → Plan + [x]/[ ]
│   │   ├── file-change      file_change completed → File Change + path
│   │   ├── hooks-deprecation error item → Config Warning
│   │   └── assistant-message agent_message → assistant text
│   ├── cursor/
│   │   └── tool-call        shellToolCall completed line
│   ├── opencode/            type=text|tool_use
│   │   ├── text
│   │   ├── tool-bash
│   │   ├── skill-input
│   │   └── todowrite-input
│   ├── pi/                  pi_types wire events
│   │   ├── message-update
│   │   ├── tool-exec-start
│   │   └── tool-exec-end
│   ├── generic/
│   │   └── assistant        type=assistant fallback
│   └── reject/              no parse / errors
│       ├── invalid-json
│       └── unrecognized-type
├── parse-messages/          []string + createdAt → []Message (traceview)
│   ├── codex/
│   │   ├── plan-and-file-change
│   │   └── hooks-warning
│   ├── cursor/
│   │   └── merge-lifecycle
│   └── edge/
│       ├── empty-lines
│       └── skip-unparseable
├── thin-wrapper/            agent_trace backward compat
│   ├── aliases-resolve
│   ├── parse-reexport
│   └── adapter-registration
├── print-integration/       print.FormatTraceLine (no agent_trace adapters)
│   ├── format-trace-line-opencode
│   └── format-trace-line-codex
└── summary/                 summary package helpers
    ├── compact-output-truncates
    └── title-from-identifier
```

## Test Leaves

| Leaf | Mode | Description |
|------|------|-------------|
| `parse-line/codex/plan-todo` | parse_line | todo_list updated → Plan activity + checkbox summary |
| `parse-line/codex/file-change` | parse_line | file_change → File Change + FileChanges |
| `parse-line/codex/hooks-deprecation` | parse_line | codex_hooks deprecation → Config Warning |
| `parse-line/codex/assistant-message` | parse_line | agent_message → assistant Message |
| `parse-line/cursor/tool-call` | parse_line | completed shellToolCall → Shell + call_id |
| `parse-line/opencode/text` | parse_line | text part → assistant Message |
| `parse-line/opencode/tool-bash` | parse_line | bash tool_use → Shell + command in summary |
| `parse-line/opencode/skill-input` | parse_line | Skill input → name + arguments in summary |
| `parse-line/opencode/todowrite-input` | parse_line | TodoWrite → status:content in summary |
| `parse-line/pi/message-update` | parse_line | message_update delta → assistant Message |
| `parse-line/pi/tool-exec-start` | parse_line | tool_execution_start → in_progress activity |
| `parse-line/pi/tool-exec-end` | parse_line | tool_execution_end → completed activity |
| `parse-line/generic/assistant` | parse_line | generic assistant message |
| `parse-line/reject/invalid-json` | parse_line | malformed JSON → not ok |
| `parse-line/reject/unrecognized-type` | parse_line | JSON with no adapter → not ok |
| `parse-messages/codex/plan-and-file-change` | parse_messages | 3 lines → 2 merged tool-call messages |
| `parse-messages/codex/hooks-warning` | parse_messages | hooks deprecation → 1 Config Warning |
| `parse-messages/cursor/merge-lifecycle` | parse_messages | started+completed → 1 call, FinishedAt set |
| `parse-messages/edge/empty-lines` | parse_messages | no input lines → empty slice |
| `parse-messages/edge/skip-unparseable` | parse_messages | garbage lines skipped |
| `thin-wrapper/aliases-resolve` | thin_wrapper | Message alias matches traceview type |
| `thin-wrapper/parse-reexport` | thin_wrapper | ParseMessages delegates correctly |
| `thin-wrapper/adapter-registration` | thin_wrapper | blank import registers opencode parser |
| `print-integration/format-trace-line-opencode` | print | opencode text → ASSISTANT output |
| `print-integration/format-trace-line-codex` | print | codex item → ASSISTANT via fallback |
| `summary/compact-output-truncates` | summary | long output omits middle |
| `summary/title-from-identifier` | summary | snake_case → Title Case |

## How to Run

```sh
doctest vet ./agent/event/tests/traceparse
doctest test ./agent/event/tests/traceparse
```

```go
import (
	"encoding/json"
	"reflect"
	"testing"

	agent_trace "github.com/xhd2015/agent-pro/agent_trace"
	legacytypes "github.com/xhd2015/agent-pro/agent_trace/types"
	"github.com/xhd2015/agent-pro/agent/event/print"
	"github.com/xhd2015/agent-pro/agent/event/summary"
	"github.com/xhd2015/agent-pro/agent/event/traceparse"
	traceview "github.com/xhd2015/agent-pro/agent/event/traceview"

	_ "github.com/xhd2015/agent-pro/agent/event/traceparse"
)

type Request struct {
	Mode      string
	SubMode   string
	RawLine   string
	RawLines  []string
	CreatedAt string
	LongOutput string
	Identifier string
}

type Response struct {
	Output  string
	OK      bool
	AliasOK bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "parse_line":
		parsed, ok := traceparse.ParseTraceLine([]byte(req.RawLine))
		data, _ := json.Marshal(parsed)
		return &Response{Output: string(data), OK: ok}, nil
	case "parse_messages":
		msgs := traceview.ParseMessages(req.RawLines, req.CreatedAt)
		data, _ := json.Marshal(msgs)
		return &Response{Output: string(data)}, nil
	case "thin_wrapper":
		return runThinWrapper(t, req)
	case "print":
		out := print.FormatTraceLine(req.RawLine)
		return &Response{Output: out}, nil
	case "summary":
		return runSummary(req), nil
	default:
		t.Fatalf("unknown Mode %q", req.Mode)
		return nil, nil
	}
}

func runThinWrapper(t *testing.T, req *Request) (*Response, error) {
	switch req.SubMode {
	case "aliases":
		aliasOK := reflect.TypeOf(agent_trace.Message{}) == reflect.TypeOf(traceview.AgentTraceMessage{})
		return &Response{AliasOK: aliasOK}, nil
	case "parse_reexport":
		lines := req.RawLines
		if len(lines) == 0 {
			lines = []string{
				`{"type":"item.completed","item":{"id":"item_0","type":"error","message":"` + "`[features].codex_hooks` is deprecated. Use `[features].hooks` instead." + `"}}`,
			}
		}
		createdAt := req.CreatedAt
		if createdAt == "" {
			createdAt = "2026-05-25T18:26:22.524536+08:00"
		}
		msgs := agent_trace.ParseMessages(lines, createdAt)
		data, _ := json.Marshal(msgs)
		return &Response{Output: string(data)}, nil
	case "adapter_registration":
		_ = agent_trace.Message{}
		line := req.RawLine
		if line == "" {
			line = `{"type":"text","part":{"type":"text","text":"registered"}}`
		}
		parsed, ok := legacytypes.ParseAgentTraceLine([]byte(line))
		data, _ := json.Marshal(parsed)
		return &Response{Output: string(data), OK: ok}, nil
	default:
		t.Fatalf("unknown thin_wrapper SubMode %q", req.SubMode)
		return nil, nil
	}
}

func runSummary(req *Request) *Response {
	switch req.SubMode {
	case "compact":
		return &Response{Output: summary.CompactTraceOutput(req.LongOutput)}
	case "title":
		return &Response{Output: summary.TitleFromIdentifier(req.Identifier)}
	default:
		return &Response{}
	}
}
```