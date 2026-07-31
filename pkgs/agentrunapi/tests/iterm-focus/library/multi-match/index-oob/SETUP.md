# Scenario

**Feature**: out-of-range index errors without focusing

```
Index=99 + 2 candidates -> error; FocusITerm never
```

## Steps

1. Set Index far past candidate count.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	idx := 99
	req.Index = &idx
	req.DryRun = false
	return nil
}
```
