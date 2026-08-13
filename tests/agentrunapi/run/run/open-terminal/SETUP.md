# Scenario

**Feature**: OpenTerminal builds --open follow-up and prefixes AGENT_RUN_HOME

```
Run(OpenTerminal=true, StoreHome=temp, OpenFn spy, Wait hook)
  -> follow-up has --open, AGENT_RUN_HOME=, no --new-terminal, no --detach
```

## Steps

1. Do not install Launch (production open path). Install Wait no-op.
2. Set StoreHome to a temp dir.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.OpenTerminal = true
	req.InstallWait = true
	req.StoreHome = t.TempDir()
	req.SessionID = "run-open-1"
	return nil
}
```
