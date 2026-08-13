# Scenario

**Feature**: default Run uses detach (no new terminal)

```
Run(OpenTerminal=false, Launch+Wait hooks) -> LaunchOpenTerminal=false
```

## Steps

1. Install Launch + Wait hooks; leave OpenTerminal false.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.OpenTerminal = false
	return nil
}
```
