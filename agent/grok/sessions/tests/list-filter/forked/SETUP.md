# Scenario

**Feature**: Forked=true keeps only forked sessions

```
Forked=true
  -> session_kind ∈ {fork, subagent_fork} OR forked_at non-empty non-whitespace
  -> plain main / plain subagent without forked_at excluded
```

## Preconditions

- Forked independent of role unless both set (combos under combo/).
- Whitespace-only forked_at is not forked.

## Steps

1. Seed forked and non-forked fixtures.
2. Forked=true.
3. Assert survivor IDs (and Kind where useful).
