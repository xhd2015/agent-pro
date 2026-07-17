# Scenario

**Feature**: Classify maps store + probe lifecycle → Mode

```
Store.GetSession / missing
  + Probe(ResumeReady, RunnerExited?)
  -> ModeRun | ModeSend | ModeResume + found
```

## Preconditions

- Injectable `Probe` when live/resume/else need lifecycle without TTY.
- Missing session does not call probe (or probe nil is fine).

## Steps

1. Set harness mode `classify`.
2. Leaf seeds meta and/or probe script as needed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "classify"
	return nil
}
```
