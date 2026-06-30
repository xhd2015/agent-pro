# Scenario

**Feature**: list queries daemon session registry

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list"}
	return nil
}
```