# Scenario

**Feature**: the session block counts what each provider has on disk, per tree shape

```
grok        <home>/sessions/**/summary.json          created_at / updated_at
codex       <home>/sessions/YYYY/MM/DD/rollout-*.jsonl   name start / file mtime
commandcode <home>/projects/<project>/<uuid>.jsonl       first record / file mtime
doctest <- SessionsBlock{total, oldest, newest}
```

## Preconditions

- The usage fetch keeps its default fixtures, so only the session tree varies.
- `total` counts every session on disk, children included: subagent and fork
  sessions are real usage of the account.
- `oldest` is the earliest session start and `newest` the latest recorded
  activity; both are UTC.

## Steps

1. Leaf seeds one provider's tree with the root's `Write*` helpers.
2. Run collects; Assert reads that provider's `SessionsBlock`.

## Context

- A tree that cannot be read is reported through `sessions.error`; a tree that is
  simply absent reports total 0 with no error.
