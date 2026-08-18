# Scenario

**Feature**: resolve errors for missing flag, empty id, not-found, ambiguous

```
RunSessions(["resolve", ...])
  -> non-nil error with library/CLI message
  -> stdout empty (no JSON error body required)
```

## Preconditions

- Assert returned error text (no `agent-run: ` prefix).
- Omitted flag vs empty flag are distinct messages.
