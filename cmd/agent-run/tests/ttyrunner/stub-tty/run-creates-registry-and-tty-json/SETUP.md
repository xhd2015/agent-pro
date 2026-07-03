# Scenario

**Feature**: stub-tty run creates registry and tty.json

```
stub-tty run -> stub-tty-registry + sessions/stub-tty/.../tty.json
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "run-creates-registry-and-tty-json"
	return nil
}
```
