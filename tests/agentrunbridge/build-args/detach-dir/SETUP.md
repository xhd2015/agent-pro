# Scenario

**Feature**: detach-profile argv with workspace dir and allow-relocate

```
BuildArgs(... Detach + WorkspaceDir + AllowRelocateResumeSessionDir)
  -> … --dir=<ws> --allow-relocate-resume-session-dir --detach -- <prompt>
```

## Preconditions

- Same detach-profile base as `detach-auto-send`.
- Non-empty `WorkspaceDir`; `AllowRelocateResumeSessionDir=true`.
- No `NoSubmit` (default detach does not inject-only).

## Steps

1. Set detach-profile flags plus dir and allow-relocate.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-detach-dir"
	req.Prompt = "detach with dir"
	req.AgentRunner = "grok-tty"
	req.WorkspaceDir = "/tmp/ws-detach"
	req.AllowRelocateResumeSessionDir = true
	req.AutoSendOrResume = true
	req.NewTerminal = false
	req.Open = false
	req.Detach = true
	return nil
}
```
