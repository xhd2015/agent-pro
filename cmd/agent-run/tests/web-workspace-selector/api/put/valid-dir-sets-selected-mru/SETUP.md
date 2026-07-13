# Scenario

**Feature**: PUT valid directory sets selected workspace and heads MRU (A2)

```
PUT /workspace {"path": SelectPath}
  -> 200
  -> selected_workspace = SelectPath
  -> recent_workspaces[0] = SelectPath
```

## Preconditions

- Existing absolute directory.
- Expect RED until endpoint exists.

## Steps

1. Create select dir; start web; PUT once.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "valid-dir-sets-selected-mru"
	req.SelectPath = makeSelectDir(t, req, "valid-ws")
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "put", Method: "PUT", Path: "/api/agent-run/workspace", Body: putWorkspaceBody(req.SelectPath)},
	}
	return nil
}
```
