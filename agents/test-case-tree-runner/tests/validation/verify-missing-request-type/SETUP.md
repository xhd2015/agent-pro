## Steps
- Point InputDir to the missing-request-type fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "missing-request-type")
	return nil
}
```
