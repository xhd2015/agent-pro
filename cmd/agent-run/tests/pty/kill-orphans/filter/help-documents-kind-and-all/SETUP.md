# Scenario

**Feature**: `pty kill-orphans --help` documents `--kind` and `--all`

```
agent-run pty kill-orphans --help -> --kind, --all
```

## Steps

1. Run kill-orphans help (no spawn).
2. Assert `--kind` and `--all` are documented; exit 0; trailing newline.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "" // plain CLI exec, no spawn
	req.Args = []string{"pty", "kill-orphans", "--help"}
	return nil
}
```
