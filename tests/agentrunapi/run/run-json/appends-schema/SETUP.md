# Scenario

**Feature**: RunJSON appends schema example and abs temp path outside worktree

```
RunJSON -> LaunchPrompt contains schema + ResultFile; ResultFile not under WorkspaceDir
```

## Steps

1. Launch + Wait write a stub JSON so RunJSON can return.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallLaunch = true
	req.InstallWait = true
	req.WaitWriteJSON = `{"ok":true}`
	return nil
}
```
