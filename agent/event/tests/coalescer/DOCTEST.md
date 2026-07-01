# Coalescer Tests

## Version

0.0.2

# DSN (Domain Specific Notion)

This tree tests the **`Coalescer`** in `agent/event/print`, a stateful filter
that suppresses redundant `ActionMessage` phase events on the canonical
`AgentEvent` stream.

Participants and behaviors:

- **Coalescer** — a struct with `ShouldSkip(event AgentEvent) bool`. It tracks
  the last seen `ActionMessage` `ID` and whether that ID's content has already
  been "shown" to the user.
- **Phase semantics** — an `ActionMessage` carries a `Phase` of `PhaseStart`,
  `PhaseUpdate`, `PhaseInstant`, or `PhaseEnd`. Start/update/instant phases are
  deltas that *display* content and mark the ID as shown; `PhaseEnd` is the
  terminal "full message" event that is redundant once the ID was shown.
- **Skip rule** — `ShouldSkip` returns `true` for a `PhaseEnd` whose `ID` was
  already shown by a prior start/update/instant in the same uninterrupted run.
  A `PhaseEnd` with a fresh (or previously-unseen) ID is displayed.
- **Duplicate-end handling** — two consecutive `PhaseEnd` events with the *same*
  ID skip the second; with *different* IDs both display.
- **Reset on interruption** — any non-`ActionMessage` event (e.g. `ActionError`,
  `ActionToolCall`) passes through and resets coalescer state, so a subsequent
  `PhaseEnd` with a repeated ID is displayed again.
- **Empty text** — the skip rule is keyed on phase and ID, not on text content;
  empty-text starts still mark an ID as shown, so an empty `PhaseEnd` is skipped
  and an empty `PhaseStart` still suppresses a later non-empty `PhaseEnd`.

The `Run` function in the Go block below feeds `req.Events` through a fresh
`Coalescer` one at a time and records each `ShouldSkip` boolean in
`resp.Skipped`.

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
