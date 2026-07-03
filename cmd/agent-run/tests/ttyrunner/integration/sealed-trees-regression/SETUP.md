# Scenario

**Feature**: sealed tty trees regression guard (doc pointer)

```
doctest test tty/ + grok-tty/ + codex-tty/ -> unchanged pass
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "integration"
	req.Action = "sealed-trees-regression"
	return nil
}
```
