# Scenario

**Feature**: AutoSendOrResume validates then classifies and dispatches

```
Opts -> validate (session id, Open/Detach mutex)
    -> Classify
    -> RunSession | SendLive | ResumeSession
```

## Preconditions

- Validation errors occur before dispatch hooks fire.
- Dispatch leaves install hooks and inject Probe; no agent-run binary.

## Steps

1. Set harness mode `auto`.
2. Child groups set validation vs dispatch fixtures.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "auto"
	return nil
}
```
