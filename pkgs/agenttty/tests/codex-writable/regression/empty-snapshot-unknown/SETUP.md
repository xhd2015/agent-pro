# Scenario

**Feature**: empty snapshot yields unknown writable state

```
zero-length snapshot text
  -> CheckWritable
  -> ready=false, state=unknown, reason=no terminal output
```

## Preconditions

- Fixture `codex-empty-snapshot.txt` is zero bytes.

## Steps

1. Set `req.FixtureFile` to the empty snapshot fixture.

## Context

- F5 boot/empty guard; distinct from idle-with-prompt and loading outcomes.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureEmptySnapshot
	return nil
}
```
