# Scenario

**Feature**: variant "Unable to write index file" (without "new") is transient

```
IsTransientIndexError("error: Unable to write index file", nil) -> true
```

## Steps

1. Set `ClassifyOutput` to the shorter write-index variant.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClassifyOutput = "error: Unable to write index file"
	req.WantTransient = true
	return nil
}
```
