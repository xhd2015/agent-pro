## Steps
- Point InputDir to the missing-assert fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "missing-assert")
	return nil
}
```
