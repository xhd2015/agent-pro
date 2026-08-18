# Scenario

**Feature**: resolve verb looks up by --grok-session-id and prints session id

```
seed SessionMeta(s)
  -> RunSessions(["resolve", ...flags...], store, stdout, stderr)
  -> human session_id | JSON object | error
```

## Preconditions

- First arg `resolve` is reserved and handled before bare session-id paths.
- Lookup is `FindByGrokSessionID` only; no create-on-miss.
