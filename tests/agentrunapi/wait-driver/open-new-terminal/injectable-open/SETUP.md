# Scenario

**Feature**: OpenInNewTerminal records dir and built follow-up via hook

```
OpenInNewTerminal(
  WorkspaceDir=/tmp/ws-iterm,
  FollowUpOpts open profile,
  OpenTerminal=spy)
  -> OpenCalls=1, OpenDir=ws, OpenFollowUp has auto-send + open, no --new-terminal
```

## Steps

1. Leave FollowUpLine empty so library builds from FollowUpOpts.
2. Harness installs OpenTerminal spy.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FollowUpLine = ""
	req.WorkspaceDir = "/tmp/ws-iterm"
	req.SessionID = "sess-iterm-1"
	req.Prompt = "force new"
	req.AgentRunner = "grok-tty"
	req.Open = true
	req.DriverBinary = ""
	return nil
}
```
