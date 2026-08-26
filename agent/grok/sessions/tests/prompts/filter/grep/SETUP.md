# Scenario

**Feature**: `--grep` keeps only prompts whose Text matches (case-insensitive literal)

```
# GrepSet + patterns P1, P2, …
prompts -> keep if Text contains every Pi (AND on same prompt)
  session with zero survivors -> skipped (multi); limit counts survivors only
```

## Preconditions

- GrepSet=true and Grep non-empty `[]string` for happy paths.
- Does not search title/cwd/assistant.

## Steps

1. Seed sessions with mixed matching and non-matching user prompts.
2. Set Grep / GrepSet (and optional recent/limit).
3. Assert kept texts and session membership.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.GrepSet = true
	return nil
}
```
