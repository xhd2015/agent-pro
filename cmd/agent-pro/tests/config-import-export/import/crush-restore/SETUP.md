## Preconditions
- A valid zip file with crush entries exists at the zip path.
- The operation has been set to "import".

## Steps
1. Create a zip file containing crush config entries.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "import"
	req.Agent = "crush"
	createZip(t, req.ZipPath, map[string]string{
		"crush/config/crush.json": `{"provider":"anthropic","api_key":"sk-crush"}`,
		"crush/data/crush.json":   `{"projects":["my-project"]}`,
	})
	return nil
}
```
