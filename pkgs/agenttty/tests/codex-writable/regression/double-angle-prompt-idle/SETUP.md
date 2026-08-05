# Scenario

**Bug**: Codex 0.146.0 main-chat prompt `»` (U+00BB) must be idle/ready like legacy `›`

```
snapshot has usage-limit bullet + main chat » (no ›)
  -> CheckWritable returns ready=true, state=idle
  -> DetectScreenStatus is not unknown (banner|idle, same class as ›-only idle)
```

## Preconditions

- Fixture `codex-double-angle-prompt-idle.txt` models the SeaTalk incident
  (`• You have 2 usage limit resets available` + `» Explain this codebase`).
- Current product only keys off `›` / U+203A → RED until implementer accepts `»`.

## Steps

1. Set `req.FixtureFile` to the double-angle idle fixture.

## Context

- F6 regression for Codex 0.146 composer glyph; must not wait 60s for ready.
- Does not contain legacy `›` so a ›-only heuristic cannot pass by accident.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureDoubleAngleIdle
	return nil
}
```
