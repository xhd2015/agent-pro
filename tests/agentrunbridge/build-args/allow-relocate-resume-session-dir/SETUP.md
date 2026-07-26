# Scenario

**Feature**: BuildArgs / InteractiveOpen emit `--allow-relocate-resume-session-dir`

```
RunOpts{… open profile + AllowRelocateResumeSessionDir: true}
  -> argv contains --allow-relocate-resume-session-dir
```

## Steps

1. Set open-profile flags and `AllowRelocateResumeSessionDir`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "build_args"
	req.SessionID = "sess-relocate-1"
	req.Prompt = "relocate ok"
	req.AgentRunner = "grok-tty"
	req.WorkspaceDir = "/tmp/ws-new"
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	req.AllowRelocateResumeSessionDir = true
	return nil
}
```
