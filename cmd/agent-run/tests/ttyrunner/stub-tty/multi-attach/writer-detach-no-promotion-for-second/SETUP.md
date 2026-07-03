# Scenario

**Feature**: writer detach does not promote second attach

```
writer detach -> writeClaimed stays true -> second still observer
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "writer-detach-no-promotion-for-second"
	return nil
}
```
