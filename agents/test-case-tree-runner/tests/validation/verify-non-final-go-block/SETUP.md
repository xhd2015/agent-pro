## Steps
- Point InputDir to the non-final-go-block fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join(DOCTEST_ROOT, "testdata", "non-final-go-block")
	return nil
}
```
