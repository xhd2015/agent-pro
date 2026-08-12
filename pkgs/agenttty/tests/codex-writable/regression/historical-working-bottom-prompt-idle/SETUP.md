# Scenario

**Bug**: historical `• Working` / `esc to interrupt` in scrollback must not block idle when a settled bottom `›` prompt is present

```
snapshot has earlier "• Working" + "esc to interrupt" + later "Worked for …" + main chat ›
  -> CheckWritable returns ready=true, state=idle
```

Crime scene: scorer WaitDone / evaluation hang after `score.json` — full-scrollback busy rule
fires on historical working markers even though the turn finished and a fresh `›` is visible.
Transcripts: `~/.sandbox/transcripts/*experiment-tty-idle*`, `*scorer-waitdone*`.

## Preconditions

- Fixture `codex-historical-working-bottom-prompt-idle.txt` (synthetic from crime-scene
  scrollback shape: tool run, historical Working, settled Worked-for footer, bottom `›`).
- Desired product: busy markers apply only to the **live** turn / prompt tail, not whole history.
- Current product matches `•` + `working`/`esc to interrupt` on **entire** scrollback → **RED**.

## Steps

1. Set `req.FixtureFile` to the historical-working + bottom-prompt fixture.

## Context

- F8 regression for post-turn idle false-negative (evaluation scorer WaitDone).
- Fixture **must** keep historical Working **and** a bottom `›` so a Working-only busy case
  cannot pass by accident, and an idle-only fixture cannot mask the bug.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureHistoricalWorkingBottomPromptIdle
	return nil
}
```
