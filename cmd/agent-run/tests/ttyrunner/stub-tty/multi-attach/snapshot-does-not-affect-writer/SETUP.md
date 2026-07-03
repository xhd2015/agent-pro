# Scenario

**Feature**: snapshot probe does not claim write token

```
snapshot attach -> ephemeral; writer retains unified write
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "snapshot-does-not-affect-writer"
	return nil
}
```
