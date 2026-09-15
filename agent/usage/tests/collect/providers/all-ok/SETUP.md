# Scenario

**Feature**: all three providers answer, so the cycle produces three healthy records

```
fixtures + auth + fake API  ->  Collect  ->  [grok ok, codex ok, commandcode ok]
each record also carries the sessions found in that provider's home
```

## Preconditions

- Inherits the root fixtures, credentials and fake Command Code API.
- Inherits the group's one-session-per-provider homes.

## Steps

1. Change nothing; run the cycle as configured.
