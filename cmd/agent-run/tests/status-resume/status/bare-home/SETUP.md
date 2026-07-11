# Scenario

**Feature**: bare `status` still prints isolated home path (compat)

```
empty AGENT_RUN_HOME -> agent-run status -> home: <path>\n
```

## Steps

1. Run `agent-run status` with no session id against empty home.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"status"}
	return nil
}
```
