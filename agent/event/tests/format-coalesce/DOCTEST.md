# Format-Coalesce Integration Tests

These doc-style tests verify that `FormatTraceLine` (from `agent/event/print`) correctly
integrates with the `Coalescer` to skip redundant `PhaseEnd` messages during formatted output.

Each test feeds raw JSON lines through a coalescer-augmented formatting pipeline and
checks which lines produce output vs which are suppressed.

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
