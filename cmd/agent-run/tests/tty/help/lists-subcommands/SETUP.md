# Scenario

**Feature**: `tty --help` lists all tty subcommands

```
agent-run tty --help -> status, attach, send
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Leaf `Setup` sets `req.Args` to `tty --help`.
2. `Run` executes `agent-run tty --help`.
3. `Assert` checks exit code 0 and lists status, attach, send.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"tty", "--help"}
	return nil
}
```
