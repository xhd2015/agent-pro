# Scenario

**Feature**: `--exclude` drops prompts whose Text matches (same matcher as grep)

```
# ExcludeSet + pattern Q
kept after optional grep -> drop if findLiteralCI(Text, Q)
```

## Preconditions

- ExcludeSet=true with non-empty Exclude for happy paths.
- When combined with grep: keep if match grep AND not match exclude.

## Steps

1. Seed prompts that partially match exclude/grep.
2. Set Exclude / Grep flags.
3. Assert remaining texts.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.ExcludeSet = true
	return nil
}
```
