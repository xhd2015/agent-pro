# Scenario

**Feature**: Limit applies after all filters

```
# survivors first, then cap
filters -> sort -> Limit (default 20 when <=0; max 100)
```

## Preconditions

- Leaves produce more survivors than Limit to observe clipping.
- Other filters may be set to define the survivor set.

## Steps

1. Seed many matching sessions with distinct last_active.
2. Set Limit (or leave 0 for default).
3. Assert count and newest-first order of clipped result.
