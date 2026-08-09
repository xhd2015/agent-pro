# Scenario

**Feature**: MainAgent / SubAgent role-class filters on ListWithOptions

```
# MainAgent -> keep main-agent class only
# SubAgent  -> keep sub-agent class only
# neither   -> all sessions; Kind still populated
ListWithOptions(MainAgent|SubAgent) -> role-filtered []Session
```

## Preconditions

- Role classification uses summary `session_kind` and `parent_session_id`.
- Sub-agent class: kind ∈ {subagent, subagent_resume, subagent_fork}
  OR (kind empty/absent AND parent_session_id non-empty).
- Main-agent class: complement (includes kind=fork and empty kind with no parent).
- Forked is independent (not set in this branch).

## Steps

1. Seed multi-kind fixtures under synthetic GROK_HOME.
2. Set MainAgent and/or leave both false (SubAgent under role/sub).
3. Assert survivor IDs and Kind tokens.
