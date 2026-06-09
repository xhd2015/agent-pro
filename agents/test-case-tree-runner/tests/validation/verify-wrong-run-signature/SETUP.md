## Steps
- Point InputDir to the wrong-run-signature fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "wrong-run-signature")
	return nil
}
```
