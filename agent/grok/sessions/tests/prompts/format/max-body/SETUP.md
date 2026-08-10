# Scenario

**Feature**: FormatPromptsOptions MaxBody soft-caps body to N runes + `…`

```
# MaxBodySet + MaxBodyRunes=N (N >= 1)
collapse whitespace -> softTruncateRunes(body, N) + "…" outside N
# MaxBodySet with N < 1 -> Write* clear error
```

## Preconditions

- Format-level only (`FormatPromptsOptions`); no CLI spawn.
- `MaxBodySet` true means opt-in soft-cap; `MaxBodyRunes` is **runes**, not bytes.
- Ellipsis is Unicode `…` (U+2026) **suffix outside** the N content runes.
- Invalid N (0 or negative) when set → error mentioning max-body and/or ≥ 1.

## Steps

1. Leaf seeds long prompt and sets MaxBody fields.
2. format-single / format-synthetic via Write*.
3. Assert truncated body or validation error.
