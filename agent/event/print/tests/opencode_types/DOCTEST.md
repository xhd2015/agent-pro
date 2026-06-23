# Opencode Trace Print Details

# DSN (Domain Specific Notion)

The compact trace printer receives native opencode JSONL trace lines. The
opencode trace adapter parses each `tool_use` line into a trace activity with a
friendly tool name and a summary. The print formatter renders that activity as
a compact terminal block: one header naming the tool and one or more detail
lines containing the useful input or output that explains what happened.

For opencode tools with structured inputs, the adapter is responsible for
turning the relevant fields into a readable summary before the printer renders
the block. A bare tool header is not enough for post-run trace review.

## Version

0.0.2

## Decision Tree

```
native opencode tool_use line
├── tool=Skill
│   └── input has name + arguments        -> SKILL block includes both details
└── tool=TodoWrite
    └── input has todos[] content/status  -> TODO/PLAN block includes todo details
```

## Test Leaves

| Leaf | Description |
|---|---|
| `tool-skill-input-details` | `Skill` tool input details are visible below the `SKILL` header. |
| `tool-todowrite-input-details` | `TodoWrite` todos are visible below the todo/plan header. |

## How to Run

```sh
doctest vet ./agent/event/print/tests/opencode_types
doctest test ./agent/event/print/tests/opencode_types/...
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/event/print"
)

type Request struct {
	Line string
}

type Response struct {
	Output string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return &Response{Output: print.FormatTraceLine(req.Line)}, nil
}
```
