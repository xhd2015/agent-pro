## Steps
- Point InputDir to the missing-leaf-setup fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "missing-leaf-setup")
	return nil
}
```
