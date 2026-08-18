# Scenario

**Feature**: Find/List cardinality and runner filtering by runner_session_id

```
seed SessionMetas
  -> FindByGrokSessionID / ListByRunnerSessionID(query, runners...)
  -> unique meta | not-found | empty-id | ambiguous | filtered list
```

## Preconditions

- Find is List filtered to `grok` + `grok-tty` then cardinality rules.
- Empty query after trim uses the CLI empty-id message for both Find and List.
- Cache may populate as a side effect; asserts here focus on lookup outcomes.

## Steps

1. Leaf seeds metas and sets `Op` to `find`, `list`, or `find_and_list`.
2. `Run` creates sessions then calls the product APIs.
3. Assert checks SessionID / list length / exact or substring errors.
