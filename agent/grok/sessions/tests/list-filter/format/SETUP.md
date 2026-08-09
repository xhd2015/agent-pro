# Scenario

**Feature**: FormatListTable / FormatListTableWithHits always show KIND column

```
ListWithOptions -> []Session with Kind
  -> FormatListTable / FormatListTableWithHits
  -> header: SESSION ID  KIND  LAST ACTIVE  TITLE  MSGS  CWD
  -> empty list still "No sessions found" (no header)
```

## Preconditions

- KIND column always present when rows exist (even without role flags).
- Column order: SESSION ID then KIND then LAST ACTIVE …
- Display tokens: main | sub | sub+ | sub-f | fork

## Steps

1. Seed sessions (or leave empty).
2. WantFormat or WantFormatHits on Request.
3. Assert header order, tokens, or empty phrase.
