# Scenario

**Feature**: zero Timeout becomes 30 minutes

```
Run(Timeout=0, Launch hook) -> LaunchTimeout == DefaultRunTimeout
```

## Steps

1. UseZeroTimeout so harness leaves Timeout at 0.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.UseZeroTimeout = true
	return nil
}
```
