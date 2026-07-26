# Scenario

**Feature**: RunInteractiveOpen launch argv equals BuildArgs of filled RunOpts

```
filled = InteractiveOpen mapping -> RunOpts
BuildArgs(filled) == launch argv from RunInteractiveOpen  # single code path
```

## Preconditions

- Spy via fake RunCommand captures launch argv.
- ExpectedArgs computed in harness Mode `interactive_open_vs_run`.

## Steps

1. Set Mode `interactive_open_vs_run`.
2. Provide session, prompt, optional dir/nosubmit for a non-trivial fill.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "interactive_open_vs_run"
	req.SessionID = "sess-io-compose"
	req.Prompt = "compose check"
	req.WorkspaceDir = "/compose/ws"
	req.NoSubmit = true
	req.AgentRunner = "" // default grok-tty in fill
	return nil
}
```
