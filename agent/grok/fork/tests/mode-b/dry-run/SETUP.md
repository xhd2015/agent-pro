# Scenario

**Feature**: Mode B `--dry-run` prints current-terminal plan and does not launch

```
fork.Main(["--session-id", id, "--dry-run"])
  -> stdout locked plan; terminal current
  -> no OpenInNewTerminal / RunForeground
```

## Steps

1. Args `--session-id` + `--dry-run`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--session-id", fixtureSessionID, "--dry-run"}
	return nil
}
```
