# Coalescer Tests

These doc-style tests verify the `agent/event/print.Coalescer` struct and its
`ShouldSkip(event AgentEvent) bool` method. The coalescer statefully tracks
consecutive `ActionMessage` events by `ID` to skip redundant `PhaseEnd` messages
whose content was already shown via `PhaseStart`/`PhaseUpdate`/`PhaseInstant` deltas.

## Decision Tree

```
Level 1: Scenario category
├── standalone/         Single events — never skipped
│   ├── phase-end/      PhaseEnd alone → displayed
│   ├── phase-start/    PhaseStart alone → displayed
│   ├── phase-update/   PhaseUpdate alone → displayed
│   └── phase-instant/  PhaseInstant alone → displayed
├── skip-after-shown/   PhaseEnd skipped after a "show" phase
│   ├── start-then-end/       PhaseStart → PhaseEnd (same ID) → skip end
│   ├── update-then-end/      PhaseUpdate → PhaseEnd (same ID) → skip end
│   ├── instant-then-end/     PhaseInstant → PhaseEnd (same ID) → skip end
│   └── full-stream/          PhaseStart → PhaseUpdate → PhaseEnd → skip end only
├── duplicate-end/      Consecutive PhaseEnd handling
│   ├── same-id/              PhaseEnd(x) → PhaseEnd(x) → skip second
│   └── different-id/         PhaseEnd(a) → PhaseEnd(b) → display both
├── interrupted/        Non-ActionMessage resets coalescer state
│   ├── non-message-between/  PhaseEnd(a) → Error → PhaseEnd(a) → all displayed
│   └── tool-between/         PhaseEnd(a) → ToolCall → PhaseEnd(a) → all displayed
└── edge-cases/         Edge cases (empty text, boundary)
    ├── empty-start/          PhaseStart("") → PhaseEnd("text") → skip end
    ├── empty-end/            PhaseStart("text") → PhaseEnd("") → skip end
    └── all-empty/            All events empty → still follow coalescer rules
```

## Test Leaves

| Leaf | Description |
|---|---|
| **standalone** | |
| `standalone/phase-end` | Solo PhaseEnd with text → NOT skipped |
| `standalone/phase-start` | Solo PhaseStart → NOT skipped |
| `standalone/phase-update` | Solo PhaseUpdate → NOT skipped |
| `standalone/phase-instant` | Solo PhaseInstant → NOT skipped |
| **skip-after-shown** | |
| `skip-after-shown/start-then-end` | PhaseStart then PhaseEnd(same ID) → end skipped |
| `skip-after-shown/update-then-end` | PhaseUpdate then PhaseEnd(same ID) → end skipped |
| `skip-after-shown/instant-then-end` | PhaseInstant then PhaseEnd(same ID) → end skipped |
| `skip-after-shown/full-stream` | Start→Update→End → only End skipped |
| **duplicate-end** | |
| `duplicate-end/same-id` | Two PhaseEnd with same ID → second skipped |
| `duplicate-end/different-id` | Two PhaseEnd with different IDs → both displayed |
| **interrupted** | |
| `interrupted/non-message-between` | PhaseEnd→Error→PhaseEnd(same ID) → all displayed (state reset) |
| `interrupted/tool-between` | PhaseEnd→ToolCall→PhaseEnd(same ID) → all displayed (state reset) |
| **edge-cases** | |
| `edge-cases/empty-start` | PhaseStart("")→PhaseEnd("text") → end skipped (start was shown) |
| `edge-cases/empty-end` | PhaseStart("text")→PhaseEnd("") → end skipped (start was shown) |
| `edge-cases/all-empty` | PhaseStart("")→PhaseUpdate("")→PhaseEnd("") → end skipped |

## How to Run

```sh
doctest test ./agent/event/tests/coalescer
```

```go
import (
	"testing"

	print "github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)


type Request struct {
	Events []types.AgentEvent
}

type Response struct {
	Skipped []bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var c print.Coalescer
	skipped := make([]bool, len(req.Events))
	for i, ev := range req.Events {
		skipped[i] = c.ShouldSkip(ev)
	}
	return &Response{Skipped: skipped}, nil
}
```
