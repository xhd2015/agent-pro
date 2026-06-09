## Steps
- Point InputDir to the wrong-assert-signature fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "wrong-assert-signature")
	return nil
}
```
