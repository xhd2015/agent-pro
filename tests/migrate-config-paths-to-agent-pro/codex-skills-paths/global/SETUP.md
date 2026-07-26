# Scenario

**Feature**: codex global skill dirs leaf (moved out of dual leaf/group parent).

## Steps

1. Rely on parent Setup for HOME + TestCase.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	return nil
}
```
