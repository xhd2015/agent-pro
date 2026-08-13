# Scenario

**Feature**: ExitOnFinishTerminal injects /exit on open profile

```
Run(OpenTerminal=true, ExitOnFinishTerminal=true) -> SoftExitCalls=1
```

## Steps

1. Open path via OpenFn + Wait; set ExitOnFinishTerminal.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.OpenTerminal = true
	req.ExitOnFinishTerminal = true
	req.InstallWait = true
	req.SessionID = "run-exit-term-1"
	return nil
}
```
