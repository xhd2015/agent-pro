## Steps
- Point InputDir to the setup-stub-body fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "setup-stub-body")
	return nil
}
```
