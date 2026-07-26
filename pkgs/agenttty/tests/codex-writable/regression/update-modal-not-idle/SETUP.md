# Scenario

**Bug**: Update available modal must not be classified as idle chat prompt

```
live snapshot: "Update available!" + "› 1. Update now…" + "Skip until next version"
  + "Press enter to continue"
  -> CheckWritable must return ready=false, state=loading (not idle)
```

## Preconditions

- Fixture `codex-update-available-modal.txt` is the preserved live capture
  (`/tmp/codex-status-fixtures-for-req/update-available-modal.txt`).
- Current production code treats `›` as main prompt → `ready=true` / `state=idle` — RED before fix.

## Steps

1. Set `req.FixtureFile` to the update-available modal fixture.

## Context

- Critical FetchStatus regression (F2): `/status` must not be sent into the update picker.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureUpdateModal
	return nil
}
```
