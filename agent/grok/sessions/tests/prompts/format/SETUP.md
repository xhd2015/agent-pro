# Scenario

**Feature**: FormatPromptsText / FormatPromptsListText compact CLI contract

```
# format surface
SessionPrompts | []SessionPrompts + FormatPromptsOptions{Location:UTC, Now, MaxBody?}
  -> compact lines / headers / empty message
  -> full body default; optional MaxBody soft-cap
  -> trailing newline; no 👤 USER cards
```

## Preconditions

- Location forced to UTC for deterministic timestamps.
- Compact line layout: `[2006-01-02 15:04:05] text`
- Missing timestamp: `[—]`
- Multi header includes session id, relative last active, title, short cwd.
- Body: full collapsed text unless `MaxBodySet` (N ≥ 1 runes + `…`).

## Steps

1. Leaf prepares FS fixtures or synthetic SessionPrompts.
2. Calls format Op.
3. Assert text contract.
