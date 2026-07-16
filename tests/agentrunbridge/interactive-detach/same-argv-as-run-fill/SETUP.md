# Scenario

**Feature**: RunInteractiveDetach launch argv equals BuildArgs of filled detach RunOpts

```
filled = InteractiveDetach mapping -> RunOpts
BuildArgs(filled) == launch argv from RunInteractiveDetach  # single code path
```

## Preconditions

- Spy via fake RunCommand captures launch argv.
- ExpectedArgs computed in harness Mode `interactive_detach_vs_run`.

## Steps

1. Set Mode `interactive_detach_vs_run`.
2. Provide session, prompt, optional dir/relocate for a non-trivial fill.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "interactive_detach_vs_run"
	req.SessionID = "sess-id-compose"
	req.Prompt = "compose detach"
	req.WorkspaceDir = "/compose/detach-ws"
	req.AllowRelocateResumeSessionDir = true
	req.AgentRunner = "" // default grok-tty in fill
	return nil
}
```
