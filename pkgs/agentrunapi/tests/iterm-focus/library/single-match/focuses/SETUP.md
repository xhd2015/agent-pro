# Scenario

**Feature**: single match focuses the iTerm ref

```
DryRun=false + one candidate -> FocusSession ok; FocusITerm(ref) once
```

## Steps

1. Inherited single-match fixtures; DryRun false (default).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = false
	return nil
}
```
