# Scenario

**Feature**: `run --help` lists `--prepend-path` and `-e`/`--env`

```
agent-run run --help -> documents --prepend-path and -e/--env
```

## Steps

1. Run `agent-run run --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"run", "--help"}
	return nil
}
```
