# Scenario

**Feature**: neither MainAgent nor SubAgent → no role filter; Kind still set

```
no role flags
  -> all discovered sessions returned
  -> each Session.Kind populated from session_kind (+ default main)
```

## Preconditions

- MainAgent=false, SubAgent=false.
- Kind population is independent of role filters.

## Steps

1. Seed one fixture per Kind token.
2. List with high Limit, no role flags.
3. Assert all IDs and Kind tokens.
