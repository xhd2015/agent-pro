# Scenario

**Feature**: interactive open imports and calls agentrunapi.AutoSendOrResume

```
agent.go
  import pkgs/agentrunapi
  AutoSendOrResume(...)
```

## Steps

1. Mode `interactive_api`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "interactive_api"
	return nil
}
```
