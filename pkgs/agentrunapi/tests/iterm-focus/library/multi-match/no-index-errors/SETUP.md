# Scenario

**Feature**: multi match without index is an error

```
Index=nil + 2 candidates -> error; FocusITerm never; Find lists both
```

## Steps

1. Leave Index nil (default).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Index = nil
	req.DryRun = false
	return nil
}
```
