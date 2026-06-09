## Steps
- Point InputDir to the missing-go-block fixture (root SETUP.md is prose-only)

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "missing-go-block")
	return nil
}
```
