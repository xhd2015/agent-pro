# Scenario

**Feature**: `tty --help` and `tty <cmd> --help` display subcommand help

```
agent-run tty --help -> status, attach, send
agent-run tty status --help -> status-specific help
agent-run tty attach --help -> attach-specific help
agent-run tty send --help -> send-specific help
```

## Steps

1. Leaf `Setup` sets `req.Args` to the specific help invocation.
2. `Run` executes the CLI.
3. `Assert` checks that help output lists the expected subcommands.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "--help"}
	return nil
}
```
