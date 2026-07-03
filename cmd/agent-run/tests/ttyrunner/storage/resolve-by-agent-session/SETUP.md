# Scenario

**Feature**: ResolveByAgentSession links agent session to live registry

```
meta.terminal_session_id -> registry entry via ResolveByAgentSession
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
func Setup(t *testing.T, req *Request) error {
	req.Action = "resolve-by-agent-session"
	req.AgentSessionID = "sess_resolve_agent"
	return nil
}
```
