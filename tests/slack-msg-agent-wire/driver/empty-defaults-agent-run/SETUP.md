# Scenario

**Feature**: empty SLACK_LISTEN_AGENT_RUN / empty DriverBinary defaults to agent-run

```
# empty env / empty DriverBinary
compat fallback binary name "agent-run" documented or used in agent.go
# (library BuildFollowUp empty DriverBinary → agent-run; wire must not force a non-empty wrong binary)
```

## Steps

1. Mode `driver_empty`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "driver_empty"
	return nil
}
```
