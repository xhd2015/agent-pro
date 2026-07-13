# Scenario

**Feature**: selected workspace persists across subsequent status requests (A4)

```
# PUT selected then re-GET status
PUT /workspace {path: SelectPath} -> 200
GET /status -> workspace == SelectPath
```

## Preconditions

- Absolute existing directory `SelectPath` under TempDir.
- Feature not yet implemented: expects RED (PUT missing / no persistence).

## Steps

1. Create select dir; start web.
2. PUT workspace then GET status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "persists-after-put"
	req.SelectPath = makeSelectDir(t, req, "persist-ws")
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "put", Method: "PUT", Path: "/api/agent-run/workspace", Body: putWorkspaceBody(req.SelectPath)},
		{Name: "status", Method: "GET", Path: "/api/agent-run/status"},
	}
	return nil
}
```
