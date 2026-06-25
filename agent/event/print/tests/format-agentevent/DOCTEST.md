# FormatAgentEvent Tests

These doc-style tests verify `print.FormatAgentEvent`, the new public function
that formats `AgentEvent` structs as human-readable strings. The function maps
each `ActionType` to a distinct visual format with emoji icons.

These tests also exercise `print.FormatTraceLine`'s new AgentEvent primary path:
when `FormatTraceLine` receives a valid AgentEvent JSON line, it unmarshals and
delegates to `FormatAgentEvent`. The existing adapter fallback is covered by the
sibling `format-*` and `pi_types/` tests.

## Decision Tree

The root splits on `AgentEvent.Type` — the single most significant parameter
that determines all formatting.

```
format-agentevent/
├── DOCTEST.md            # This file
├── SETUP.md              # Root: Request/Response, Run calls FormatAgentEvent
│
├── tool-call/            # Type = tool_call
│   ├── SETUP.md          #   Sets Type=tool_call, Tool=bash
│   ├── basic/            #   output=file1.txt → RUN, file1.txt
│   ├── with-text/        #   text="Running" + output → RUN, Running, output
│   ├── with-changes/     #   Tool=write, changes[{path,kind}] → EDIT, kind path
│   ├── exit-zero/        #   exit_code=0 → no FAILED
│   ├── exit-error/       #   exit_code=1 → FAILED
│   └── all-fields/       #   text+output+changes+exit_code≠0 → all rendered
│
├── message/              # Type = message
│   ├── SETUP.md          #   Sets Type=message
│   ├── basic/            #   Text="hello world" → 💬, hello world
│   └── empty-text/       #   Text="" → 💬 (empty message still renders)
│
├── think/                # Type = think
│   ├── SETUP.md          #   Sets Type=think
│   ├── basic/            #   Text="reasoning..." → 💭, reasoning...
│   └── empty-text/       #   Text="" → 💭 (empty think still renders)
│
├── error/                # Type = error
│   ├── SETUP.md          #   Sets Type=error
│   ├── basic/            #   Text="boom" → ❌, boom, FAILED
│   └── empty-text/       #   Text="" → ❌, FAILED
│
├── step/                 # Types = step_start, step_finish
│   ├── SETUP.md          #   (no shared preconditions)
│   ├── step-start/       #   → ▶ STEP START
│   └── step-finish/      #   → ◼ STEP FINISH
│
└── default/              # Unhandled types: done, sleep, unknown
    ├── SETUP.md          #   (no shared preconditions)
    ├── done/             #   Type=done → [done] + text
    ├── sleep/            #   Type=sleep → [sleep] + text
    └── unknown/          #   Type="custom_type" → [custom_type] + text
```

## Leaf Index

### tool-call — 6 leaves
| Leaf | Type | Key Fields | Expected Contains |
|------|------|-----------|-------------------|
| `tool-call/basic` | tool_call | tool=bash, output=file1.txt | `RUN`, `file1.txt` |
| `tool-call/with-text` | tool_call | tool=bash, text=Running..., output=result | `RUN`, `Running`, `result` |
| `tool-call/with-changes` | tool_call | tool=write, changes=[{a.txt,create}] | `EDIT`, `create a.txt` |
| `tool-call/exit-zero` | tool_call | tool=bash, exit_code=0 | `RUN`, no `FAILED` |
| `tool-call/exit-error` | tool_call | tool=read, exit_code=1 | `READ`, `FAILED` |
| `tool-call/all-fields` | tool_call | tool=bash, text+output+changes+exit_code=2 | `RUN`, text, output, change, `FAILED` |

### message — 2 leaves
| Leaf | Type | Text | Expected Contains |
|------|------|------|-------------------|
| `message/basic` | message | hello world | `💬`, `hello world` |
| `message/empty-text` | message | "" | `💬` (renders even with empty text) |

### think — 2 leaves
| Leaf | Type | Text | Expected Contains |
|------|------|------|-------------------|
| `think/basic` | think | reasoning about... | `💭`, `reasoning` |
| `think/empty-text` | think | "" | `💭` (renders even with empty text) |

### error — 2 leaves
| Leaf | Type | Text | Expected Contains |
|------|------|------|-------------------|
| `error/basic` | error | boom | `❌`, `boom`, `FAILED` |
| `error/empty-text` | error | "" | `❌`, `FAILED` |

### step — 2 leaves
| Leaf | Type | Expected Contains |
|------|------|-------------------|
| `step/step-start` | step_start | `▶ STEP START` |
| `step/step-finish` | step_finish | `◼ STEP FINISH` |

### default — 3 leaves
| Leaf | Type | Text | Expected Contains |
|------|------|------|-------------------|
| `default/done` | done | all done | `[done]`, `all done` |
| `default/sleep` | sleep | waiting 5s | `[sleep]`, `waiting 5s` |
| `default/unknown` | custom_type | some text | `[custom_type]`, `some text` |

Total: **17 leaves** across **6 action-type groups**.

## How to Run

```sh
# All format-agentevent tests
doctest test -v ./agent/event/print/tests/format-agentevent

# All print package tests (existing + new)
doctest test -v ./agent/event/print/tests/...
```

```go
import (
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"

	"github.com/xhd2015/agent-pro/agent/event/print"
)


type Request struct {
	Type     types.ActionType
	Text     string
	Tool     string
	Output   string
	ExitCode *int
	Changes  []types.FileChange
}

type Response struct {
	Output string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	event := types.AgentEvent{
		Type:     req.Type,
		Text:     req.Text,
		Tool:     req.Tool,
		Output:   req.Output,
		ExitCode: req.ExitCode,
		Changes:  req.Changes,
	}
	return &Response{Output: print.FormatAgentEvent(event)}, nil
}
```
