## Steps
- Point InputDir to the child-redefines-request fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "child-redefines-request")
	return nil
}
```
