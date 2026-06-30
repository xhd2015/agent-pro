# Scenario

**Feature**: ResolveTarget maps id-or-name to a single session

```
# resolution
List sessions -> match id or name -> SessionInfo or ambiguity error
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "resolve-id"
	return nil
}
```