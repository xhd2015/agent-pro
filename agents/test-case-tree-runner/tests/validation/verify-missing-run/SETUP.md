## Steps
- Point InputDir to the missing-run fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "missing-run")
	return nil
}
```
