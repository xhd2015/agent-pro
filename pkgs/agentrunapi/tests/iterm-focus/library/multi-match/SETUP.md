# Scenario

**Feature**: multiple iTerm candidates require --index policy

```
two real TTYs -> two FocusCandidates -> Index nil | valid | OOB
```

## Steps

1. Apply `multiMatchFixtures`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	multiMatchFixtures(req)
	return nil
}
```
