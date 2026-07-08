# Scenario

**Bug**: session-18 false positive — scrollback `git working tree status` must not block idle prompt

```
snapshot scrollback contains "git working tree status" (tool output)
prompt tail shows idle ❯ + Enter:send
  -> CheckWritable must return ready=true, state=idle
```

## Preconditions

- Fixture `grok-after_git-idle-false-positive-session18-synth.txt` synthesized from session-18 report.
- Current production code false-positives as `busy` — this leaf is RED before fix.

## Steps

1. Set `req.FixtureFile` to the session-18 regression fixture.

## Context

- Primary bug reproducer (F2); send-queue drainer must not stall on git status scrollback.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FixtureFile = fixtureFalsePositiveSession18
	return nil
}
```