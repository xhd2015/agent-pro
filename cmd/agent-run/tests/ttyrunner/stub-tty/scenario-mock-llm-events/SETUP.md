# Scenario

**Feature**: scenario llm_events written to events.jsonl

```
llm_events -> events.jsonl assistant message
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "scenario-mock-llm-events"
	return nil
}
```
