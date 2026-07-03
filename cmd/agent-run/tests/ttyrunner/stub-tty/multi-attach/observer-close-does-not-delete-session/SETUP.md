# Scenario

**Feature**: observer disconnect does not delete session

```
observer WS close -> session registry remains
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "observer-close-does-not-delete-session"
	return nil
}
```
