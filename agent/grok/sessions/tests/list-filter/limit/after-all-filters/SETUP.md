# Scenario

**Feature**: Limit clips after place filter (newest survivors first)

```
four sessions under cwdA; PlaceCWDs=[A]; Limit=2
  -> two newest only
```

## Preconditions

- Place filter selects all four; Limit=2 applies after.

## Steps

1. Write four A sessions with known last_active order.
2. PlaceCWDs=[A], Limit=2.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 2
	req.PlaceCWDs = []string{absPath(t, cwdA)}
	// also a B session that would be newest overall but fails place
	writeListSession(t, req.GrokHome, idB1, atFixed(-1*time.Minute), cwdB, "B newest overall")
	writeListSession(t, req.GrokHome, idA1, atFixed(-40*time.Minute), cwdA, "A4 oldest")
	writeListSession(t, req.GrokHome, idA2, atFixed(-30*time.Minute), cwdA, "A3")
	writeListSession(t, req.GrokHome, idA3, atFixed(-20*time.Minute), cwdA, "A2")
	writeListSession(t, req.GrokHome, idD1, atFixed(-10*time.Minute), cwdA, "A1 newest in place")
	return nil
}
```
