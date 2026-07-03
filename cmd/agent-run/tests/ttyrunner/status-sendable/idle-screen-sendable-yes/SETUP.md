# Scenario

**Feature**: idle scrollback reports sendable yes

```
scrollback with prompt -> tty status --json -> sendable: true
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "idle-screen-sendable-yes"
	req.RegistryDir = "grok-tty-registry"
	req.FakePTYWrapScrollback = "GROK_TTY_BANNER\nGrok › prompt\nResponse: done\n› "
	return nil
}
```
