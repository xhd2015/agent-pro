# Scenario

**Feature**: top-level help documents `attach` subcommand

```
agent-run --help → stdout contains attach
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