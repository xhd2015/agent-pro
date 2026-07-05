# Scenario

**Feature**: top-level help documents `send` subcommand

```
agent-run --help → stdout contains send
```

## Steps

1. Run `agent-run --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```