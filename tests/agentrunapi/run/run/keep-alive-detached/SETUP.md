# Scenario

**Feature**: KeepAliveDetached skips SoftExit

```
Run(KeepAliveDetached=true) -> SoftExitCalls=0
```

## Steps

1. Detach with KeepAliveDetached true.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.OpenTerminal = false
	req.KeepAliveDetached = true
	return nil
}
```
