# Scenario

**Feature**: interactive open no longer uses agentrunbridge.RunInteractiveOpen

```
agent.go must NOT contain RunInteractiveOpen
```

## Steps

1. Mode `interactive_no_bridge_open`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "interactive_no_bridge_open"
	return nil
}
```
