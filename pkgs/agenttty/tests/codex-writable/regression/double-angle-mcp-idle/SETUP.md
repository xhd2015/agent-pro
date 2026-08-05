# Scenario

**Bug**: MCP startup incomplete + main chat `»` (no `›`) must still be idle

```
snapshot has "MCP startup incomplete" + main chat » only
  -> CheckWritable returns ready=true, state=idle
  (same rule as MCP incomplete + ›)
```

## Preconditions

- Fixture `codex-double-angle-mcp-incomplete.txt` has MCP incomplete warning and
  only the double-angle prompt glyph (no legacy `›`).
- Current product treats MCP incomplete without `›` as `loading` and never
  recognizes `»` as the main prompt → RED until implementer accepts `»` on both
  the MCP exception path and the idle-prompt path.

## Steps

1. Set `req.FixtureFile` to the double-angle MCP incomplete fixture.

## Context

- F7 MECE companion to F6: guards the separate MCP-incomplete branch that only
  checked for `›` / U+203A before the idle classification.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureDoubleAngleMCP
	return nil
}
```
