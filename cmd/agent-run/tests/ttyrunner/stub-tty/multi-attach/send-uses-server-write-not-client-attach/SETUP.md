# Scenario

**Feature**: tty send uses server WriteInput not client attach

```
writer holds client write; tty send -> server controller inject
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "send-uses-server-write-not-client-attach"
	return nil
}
```
