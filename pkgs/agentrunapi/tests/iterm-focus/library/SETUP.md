# Scenario

**Feature**: library FindITermForSession / FocusSession with injectables

```
Store + session_id + ListProcs + ListITerm + FocusITerm
  -> FocusSession -> chosen | error; FocusITerm only when unique/index + !DryRun
```

## Preconditions

- Leaves use `Phase=focus` and injectable hooks (no live iTerm).
- Store opened on isolated `req.Home` via `NewFileStore` (no env mutation).

## Steps

1. Grouping defaults `Phase=focus`.
2. Subtrees set resolve / match / policy fixtures.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "focus"
	return nil
}
```
