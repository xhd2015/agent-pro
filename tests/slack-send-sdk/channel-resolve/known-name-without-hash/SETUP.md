# Scenario

**Feature**: `general` resolves via `#general` map entry

```
resolveChannel("general") -> knownChannels["#general"] -> C0ALE44K5J6
```

## Steps

1. Args `["general"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"general"}
	return nil
}
```