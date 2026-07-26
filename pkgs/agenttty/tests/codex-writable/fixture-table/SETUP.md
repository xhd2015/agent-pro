# Scenario

**Feature**: fixture table drives full `codex-*.txt` coverage against `expectations.jsonl`

```
glob testdata/codex-writable/codex-*.txt
  -> load expectations.jsonl
  -> CheckWritable each fixture
  -> every row matches ready/state/reason-substring
```

## Preconditions

- `expectations.jsonl` has one entry per `codex-*.txt` fixture (5 rows in current seed).

## Steps

1. Set `req.RunAllFixtures = true`.
2. `Run` evaluates every manifest row.

## Context

- Primary coverage gate (F1); fails if any fixture drifts from post-fix expectations
  (update modal not idle is the critical RED row before implementer fix).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RunAllFixtures = true
	return nil
}
```
