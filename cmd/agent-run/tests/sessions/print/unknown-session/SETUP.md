# Scenario

**Feature**: print unknown session — clear failure

```
empty AGENT_RUN_HOME -> sessions missing_id --print -> exit 1
```

## Preconditions

- No session directories under home.

## Steps

1. Run `agent-run sessions missing_id --print` without seeding storage.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = printSessionArgs("missing_id")
	return nil
}
```
