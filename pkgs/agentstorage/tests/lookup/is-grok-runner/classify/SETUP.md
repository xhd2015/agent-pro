# Scenario

**Feature**: classify exact grok runners vs near-misses

```
IsGrokRunner("grok") / ("grok-tty") / ("  grok  ") -> true
IsGrokRunner("codex-tty") / ("GROK") / ("grok-foo") / ("") -> false
```

## Preconditions

- Trim applies; case and prefix variants are false.

## Steps

1. Table of true and false runners.
2. Op `is_grok`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "is_grok"
	req.RunnerCases = []RunnerCase{
		{Name: "grok", Runner: "grok", Want: true},
		{Name: "grok-tty", Runner: "grok-tty", Want: true},
		{Name: "trim-grok", Runner: "  grok  ", Want: true},
		{Name: "trim-grok-tty", Runner: " grok-tty ", Want: true},
		{Name: "codex-tty", Runner: "codex-tty", Want: false},
		{Name: "GROK-upper", Runner: "GROK", Want: false},
		{Name: "grok-foo-prefix", Runner: "grok-foo", Want: false},
		{Name: "empty", Runner: "", Want: false},
	}
	return nil
}
```
