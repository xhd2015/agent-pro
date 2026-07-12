# Scenario

**Feature**: MODE=send when session is live (`runner.exited == false`)

```
seed live bound + sendable terminal
  -> auto + prompt → enqueue/deliver (msg_N)
  -> auto + empty prompt → warn exit 0
  -> auto + --open → exit 1
```

## Steps

1. Default runner and meta status for live fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = defaultRunner
	req.MetaStatus = "running"
	// Live leaves use fake ptywrap inject; ensure hook does not replace send path.
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	return nil
}
```
