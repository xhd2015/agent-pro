# Scenario

**Feature**: --keep-tty keeps registry and tty.json alive

```
run --keep-tty -> registry + tty.json alive=true after exit
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "keep-tty-persists"
	return nil
}
```
