# Scenario

**Feature**: unique grok-family match prints session id (human or JSON)

```
seed unique grok|grok-tty with runner_session_id=UUID
  -> RunSessions resolve [--json] --grok-session-id UUID
  -> stdout session_id\n or JSON object; err nil
```

## Preconditions

- Exactly one grok-family meta for the query UUID.
