# Scenario

**Feature**: grok ancestor exists but Lsof has no session path

```
# default chain; OpenFiles empty
fork.Main([]) -> error "session not resolved"
```

## Steps

1. Clear OpenFiles so parse cannot find a uuid.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OpenFiles = map[int][]string{}
	req.Args = []string{}
	return nil
}
```
