# Scenario

**Feature**: fixture table drives full `grok-*.txt` coverage against `expectations.jsonl`

```
glob testdata/grok-writable/grok-*.txt
  -> load expectations.jsonl
  -> CheckWritable each fixture
  -> every row matches ready/state/reason
```

## Preconditions

- `expectations.jsonl` has one entry per `grok-*.txt` fixture (23 rows: historical seed + workspace-confirm + 3 modern open-ready frames).

## Steps

1. Set `req.RunAllFixtures = true`.
2. `Run` evaluates every manifest row.

## Context

- Primary coverage gate (F1); fails if any fixture drifts from manifest or implementation regresses.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunAllFixtures = true
	return nil
}
```