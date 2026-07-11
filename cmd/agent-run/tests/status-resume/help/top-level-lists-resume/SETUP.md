# Scenario

**Feature**: top-level help lists the `resume` command

```
agent-run --help -> stdout contains resume
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
