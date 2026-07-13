# Scenario

**Feature**: POST /sessions uses selected_workspace for meta.workspace (A5)

```
PUT /workspace {SelectPath}
POST /sessions {runner: fake-codex, prompt: "…"}
  -> 202/200; session.workspace == SelectPath
```

## Preconditions

- Absolute select dir.
- Expect RED until PUT + resolve-on-create exist (today uses process cwd only).

## Steps

1. Create dir; start web; PUT then POST sessions.

```go
import (
	"encoding/json"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "new-uses-selected-workspace"
	req.SelectPath = makeSelectDir(t, req, "session-ws")
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{
		"runner": "fake-codex",
		"prompt": "workspace selector session stamp",
	})
	req.HTTPSteps = []HTTPStep{
		{Name: "put", Method: "PUT", Path: "/api/agent-run/workspace", Body: putWorkspaceBody(req.SelectPath)},
		{Name: "create", Method: "POST", Path: "/api/agent-run/sessions", Body: string(body)},
	}
	return nil
}
```
