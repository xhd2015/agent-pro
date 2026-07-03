# Scenario

**Feature**: stub-tty run dual-writes registry and tty.json

```
stub-tty run -> stub-tty-registry/session-N.json + sessions/stub-tty/.../tty.json
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "dual-write-tty-json"
	return nil
}
```
