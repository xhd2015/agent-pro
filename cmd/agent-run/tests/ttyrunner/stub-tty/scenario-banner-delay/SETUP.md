# Scenario

**Feature**: scenario banner_delay_ms delays before writable

```
banner_delay_ms:800 -> run waits before completing
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "scenario-banner-delay"
	return nil
}
```
