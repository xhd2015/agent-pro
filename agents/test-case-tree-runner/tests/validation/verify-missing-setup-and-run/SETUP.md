## Steps
- Point InputDir to the missing-setup-and-run fixture (root has types only, no Setup/Run)

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "missing-setup-and-run")
	return nil
}
```
