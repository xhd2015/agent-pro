# Scenario

**Feature**: Active=true keeps only file-active sessions

```
# active_sessions.json object form
writeActiveSessions(ids) + ListWithOptions(Active=true)
  -> only listed session ids
```

## Preconditions

- Uses existing `IsFileActive` semantics.
- Descendants write active_sessions.json via writeActiveSessions.

## Steps

1. Seed sessions on disk.
2. Mark a subset active (or none).
3. Set Active=true.
