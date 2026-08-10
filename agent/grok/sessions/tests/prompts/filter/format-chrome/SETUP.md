# Scenario

**Feature**: formatter chrome for filters — omission marker, grep color, body length

```
# head/tail Omitted* -> "(...M omitted...)" line
# GrepSet + ColorMode always -> ANSI highlight on match span
# Grep + long body: full by default; window only when MaxBodySet
```

## Preconditions

- Omission marker exact shape `(...M omitted...)`.
- Color highlight when ColorMode=always and GrepSet (bold red family).
- Grep body policy: full collapsed text unless MaxBody; then window ≤ N.

## Steps

1. Seed or synthesize prompts; set format-related opts.
2. Run format-single or format-list.
3. Assert marker substring, ANSI CSI, and/or body length policy.
