# Scenario

**Feature**: Find reports not-found when no grok-family meta matches the UUID

```
seed (optional unrelated metas)
  -> FindByGrokSessionID(UUID)
  -> error: session not found: no grok session with runner_session_id …
```

## Preconditions

- Unknown UUID and empty-`runner_session_id` meta are distinct missing paths.
