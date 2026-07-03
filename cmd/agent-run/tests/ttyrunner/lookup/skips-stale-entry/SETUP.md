# Scenario

**Feature**: stale grok entry skipped; live codex returned

```
grok unreachable (removed) -> codex live entry returned
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "skips-stale-entry"
	return nil
}
```
