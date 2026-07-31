# Scenario

**Feature**: --index selects among multi candidates

```
Index=1 + 2 candidates -> focus candidate Index 1 (second, 0-based)
```

## Steps

1. Set Index to 1 (second candidate).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	idx := 1
	req.Index = &idx
	req.DryRun = false
	return nil
}
```
