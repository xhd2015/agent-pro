# Scenario

**Feature**: detach default calls SoftExit after wait

```
Run(detach zeros, SoftExit spy) -> SoftExitCalls=1
```

## Steps

1. Launch + Wait hooks; OpenTerminal false; KeepAliveDetached false.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.OpenTerminal = false
	req.KeepAliveDetached = false
	return nil
}
```
