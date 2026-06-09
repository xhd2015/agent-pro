## Steps
- Point InputDir to the wrong-setup-signature fixture

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.InputDir = filepath.Join("testdata", "wrong-setup-signature")
	return nil
}
```
