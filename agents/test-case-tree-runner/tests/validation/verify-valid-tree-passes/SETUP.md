## Steps
- Point InputDir to the valid-tree fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "valid-tree")
	return nil
}
```
