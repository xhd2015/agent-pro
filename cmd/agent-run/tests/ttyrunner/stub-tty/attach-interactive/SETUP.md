# Scenario

**Feature**: interactive attach to live stub-tty session

```
attach_mode=interactive -> writer can send input
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "attach-interactive"
	return nil
}
```
