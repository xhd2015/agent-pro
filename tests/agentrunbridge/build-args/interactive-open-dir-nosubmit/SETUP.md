# Scenario

**Feature**: open-profile argv with workspace dir and no-submit

```
BuildArgs(... + WorkspaceDir + NoSubmit)
  -> … --dir=<ws> --no-submit --open -- <prompt>
```

## Preconditions

- Same open-profile base as defaults leaf.
- Non-empty `WorkspaceDir`; `NoSubmit=true`.

## Steps

1. Set open-profile flags plus dir and no-submit.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-open-dir"
	req.Prompt = "with dir"
	req.AgentRunner = "grok-tty"
	req.WorkspaceDir = "/tmp/ws-bridge"
	req.NoSubmit = true
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	return nil
}
```
