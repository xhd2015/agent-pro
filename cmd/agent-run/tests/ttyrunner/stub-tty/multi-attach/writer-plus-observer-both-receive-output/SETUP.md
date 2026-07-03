# Scenario

**Feature**: writer and observer both receive multiplexed output

```
PTY output -> writer + observer fan-out
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "writer-plus-observer-both-receive-output"
	return nil
}
```
