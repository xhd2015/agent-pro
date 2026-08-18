# Scenario

**Feature**: unique grok-family meta resolves via Find

```
seed one grok|grok-tty with runner_session_id=UUID
  -> FindByGrokSessionID(UUID)
  -> that SessionMeta (SessionID matches)
```

## Preconditions

- Exactly one matching grok-family session for the query UUID.
- Sibling branches differ only by `meta.runner` (`grok-tty` vs legacy `grok`).
