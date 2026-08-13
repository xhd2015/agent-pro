# Scenario

**Feature**: unknown `--session-id` is not found under GROK_HOME

```
fork.Main(["--session-id", missing]) -> error "grok session not found"
```

## Steps

1. Use an id that has no `summary.json`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--session-id", "019f283b-ffff-7fff-ffff-ffffffffffff"}
	return nil
}
```
