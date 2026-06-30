# Scenario

**Feature**: PATCH renames an existing session

```
# rename flow
create -> PATCH {name} -> GET list shows new name
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "rest-rename"
	req.RenameTo = "after-rename"
	return nil
}
```