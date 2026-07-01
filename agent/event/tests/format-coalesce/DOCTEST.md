# Format-Coalesce Integration Tests

## Version

0.0.2

# DSN (Domain Specific Notion)

This tree tests the integration of **`FormatTraceLine`** (in `agent/event/print`)
with the **`Coalescer`** — i.e. that formatting raw trace JSON lines correctly
suppresses redundant `PhaseEnd` messages.

Participants and behaviors:

- **Raw JSON lines** — each test feeds a sequence of raw JSON strings (`req.Lines`),
  one per canonical `AgentEvent`. Lines may be `ActionMessage` phase events or
  non-message events such as `ActionThink`.
- **Pipeline** — for each line: trim it, skip blanks/non-JSON, unmarshal into a
  `types.AgentEvent`; if it is an `ActionMessage` and `Coalescer.ShouldSkip`
  returns `true`, suppress it (output `""`); otherwise call
  `print.FormatTraceLine(line)` and emit the formatted result.
- **Coalescer** — stateful; marks an `ActionMessage` ID as "shown" on
  start/update/instant phases and skips a subsequent `PhaseEnd` with the same ID
  whose content was already streamed via deltas.
- **Reset on non-message** — a non-`ActionMessage` line (e.g. `think`) passes
  through formatted AND resets coalescer state, so a later `PhaseEnd` that
  repeats a prior ID is displayed again rather than skipped.
- **Output contract** — `resp.Output[i]` is the formatted line for `Lines[i]`, or
  `""` when that line was suppressed by the coalescer.

The `Run` function in the Go block below is the integration pipeline that walks
`req.Lines` through the coalescer and formatter into `resp.Output`.

## Decision Tree

```
Level 1: Input sequence pattern
├── standalone-end/     A lone PhaseEnd JSON line → formatted
├── update-then-end/    PhaseUpdate JSON → PhaseEnd JSON → only update line formatted
└── think-between/      Think event between messages → think displayed, message state resets
```

## Test Leaves

| Leaf | Description |
|---|---|
| `standalone-end` | Single `PhaseEnd` JSON line → formatted output (standalone, no prior delta) |
| `update-then-end` | `PhaseUpdate` then `PhaseEnd` → only update produces output, end skipped |
| `think-between` | `PhaseEnd`(m1)→think→`PhaseEnd`(m2) → think resets state, both ends displayed |

## How to Run

```sh
doctest test ./agent/event/tests/format-coalesce
```

```go
import (
	"encoding/json"
	"strings"
	"testing"

	print "github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)


type Request struct {
	Lines []string
}

type Response struct {
	Output []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var c print.Coalescer
	output := make([]string, len(req.Lines))
	for i, line := range req.Lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
			output[i] = ""
			continue
		}
		var ev types.AgentEvent
		if err := json.Unmarshal([]byte(trimmed), &ev); err == nil && ev.Type == types.ActionMessage && c.ShouldSkip(ev) {
			output[i] = "" // suppressed by coalescer
			continue
		}
		output[i] = print.FormatTraceLine(line)
	}
	return &Response{Output: output}, nil
}
```
