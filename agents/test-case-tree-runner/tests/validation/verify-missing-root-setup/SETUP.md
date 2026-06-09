## Steps
- Point InputDir to the missing-root-setup fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "missing-root-setup")
	return nil
}
```
