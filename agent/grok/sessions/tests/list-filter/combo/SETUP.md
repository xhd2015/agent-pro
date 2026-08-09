# Scenario

**Feature**: place / recent / active / role / forked / grep filters AND together

```
# pipeline order: place -> recent -> active -> role -> forked -> grep -> sort -> limit
ListWithOptions(multi flags) -> intersection of survivors
```

## Preconditions

- Each leaf enables at least two filter dimensions.
- Place still OR within PlaceCWDs; dimensions AND across each other.
- Role×forked and place×role leaves live here alongside existing place/recent/active/grep combos.

## Steps

1. Seed multi-cwd / multi-time / active / content / kind fixtures as needed.
2. Set combination of PlaceCWDs, Recent*, Active, MainAgent/SubAgent, Forked, Grep*.
3. Assert only sessions passing every enabled filter.
