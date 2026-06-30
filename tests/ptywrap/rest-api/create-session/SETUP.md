# Scenario

**Feature**: POST creates a new terminal session

```
# REST create
POST {command, cwd, name} -> 201/200 + session id
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "rest-create"
	req.Command = []string{"sleep", "120"}
	req.Name = "rest-create-test"
	req.Cwd = absTempDir(t)
	return nil
}
```