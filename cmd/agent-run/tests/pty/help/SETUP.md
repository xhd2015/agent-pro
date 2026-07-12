# Scenario

**Feature**: help surfaces for `pty` discovery and unknown subcommand errors

```
# user discovers pty subcommands
agent-run pty --help -> lists stats, kill-orphans

# top-level help mentions pty
agent-run --help -> mentions pty

# unknown pty subcommand is rejected
agent-run pty <unknown> -> exit 1
```

## Steps

1. Leaf `Setup` sets `req.Args` for the help or error invocation.
2. `Run` executes the CLI (default mode).
3. `Assert` checks exit code and expected help/error content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping default: pty --help (leaves override).
	if len(req.Args) == 0 {
		req.Args = []string{"pty", "--help"}
	}
	return nil
}
```
