# Scenario

**Feature**: ttyrunner provider registry — builtin registration and conventions

```
ttyrunner.Register(grok-tty, codex-tty, stub-tty?) -> IDs / Get / IsTTYRunner
```

## Preconditions

- `pkgs/ttyrunner` registers grok-tty and codex-tty at init.
- `stub-tty` registers only when `AGENT_RUN_ENABLE_STUB_TTY=1`.

## Steps

1. Leaf sets `req.Operation = "registry"` and `req.Action`.
2. `Run` calls `ttyrunner.IDs`, `IsTTYRunner`, or `Get`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "registry"
	return nil
}
```
