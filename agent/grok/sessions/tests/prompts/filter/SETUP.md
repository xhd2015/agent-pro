# Scenario

**Feature**: filter pipeline for user-prompt history (grep / exclude / head|tail)

```
# pipeline after session discovery / single load
recent window? -> grep keep -> exclude drop -> head|tail slice
  -> skip empty survivors -> format with omission markers / grep highlight
```

## Preconditions

- Additive Request fields: Grep/GrepSet, Exclude/ExcludeSet, Head/HeadSet, Tail/TailSet, ColorMode.
- Zero-value fields preserve existing non-filter behavior.
- Matcher: case-insensitive literal on `UserPrompt.Text` only.
- Head and Tail mutually exclusive; N >= 1 when set; empty pattern when *Set → error.
- Omission marker `(...M omitted...)` is formatter chrome only (not a UserPrompt).

## Steps

1. Grouping narrows to filter concern; leaves set concrete opts and fixtures.
2. Run exercises list / single+FilterUserPrompts / filter / format-* as needed.
3. Assert structured keep lists, session skip, errors, and format chrome.

## Context

- Filter fixture ids: `idFilterSingle`, `idFilterHead`, `idFilterGrepA/B/C`, multiSessionID.
- Fixed Now from root.
