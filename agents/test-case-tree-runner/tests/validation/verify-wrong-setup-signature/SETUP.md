## Steps
- Point InputDir to the wrong-setup-signature fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "wrong-setup-signature")
	return nil
}
```
