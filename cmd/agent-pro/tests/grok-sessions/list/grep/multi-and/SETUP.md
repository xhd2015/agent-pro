# Scenario

**Feature**: multiple `--grep` patterns require AND on the same field/line

```
# patterns P1,P2 → a hit unit must contain both (case-insensitive literal)
ListWithGrep(patterns) -> SessionMatch only when ≥1 unit has every Pi
```

## Preconditions

- Descendants set `req.Grep` to two or more patterns.
- Color never for plain hit-line asserts.

## Steps

1. Parent leaves Grep unset; leaves set patterns + fixtures.
2. Assert keep vs drop based on same-unit vs split-unit placement.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Color == "" {
		req.Color = "never"
	}
	return nil
}
```
