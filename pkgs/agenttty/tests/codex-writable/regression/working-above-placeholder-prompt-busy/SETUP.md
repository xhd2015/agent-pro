# Scenario

**Bug**: live `• Working` immediately above a placeholder composer `›` must be busy

```
snapshot: user › + "• Working (6s • esc to interrupt)" + placeholder ›
  -> CheckWritable returns ready=false, state=busy
```

Crime scene: idle-timeout experiment snap-00 (`llm-mock-run-codex`). Last `›` is
`Write tests for @filename…`; prompt-region snapped to that line and dropped Working.

## Preconditions

- Fixture `codex-working-above-placeholder-prompt-busy.txt`.
- Historical Working **above** settled `Worked for` stays F8 idle; this leaf has **no**
  `Worked for` and live Working **below** the inject › / **above** the last ›.

## Steps

1. Set `req.FixtureFile` to the working-above-placeholder fixture.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureWorkingAbovePlaceholderBusy
	return nil
}
```
