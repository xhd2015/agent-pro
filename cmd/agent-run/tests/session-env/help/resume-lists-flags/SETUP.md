# Scenario

**Feature**: `resume --help` lists `--prepend-path` and `-e`/`--env`

```
agent-run resume --help -> documents --prepend-path and -e/--env
```

## Steps

1. Run `agent-run resume --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"resume", "--help"}
	return nil
}
```
