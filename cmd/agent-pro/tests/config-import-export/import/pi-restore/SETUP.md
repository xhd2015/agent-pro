## Preconditions
- A valid zip file with pi entries exists at the zip path.
- The operation has been set to "import".

## Steps
1. Create a zip file containing pi config entries.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "import"
	req.Agent = "pi"
	createZip(t, req.ZipPath, map[string]string{
		"pi/auth.json":    `{"token":"pi-token"}`,
		"pi/settings.json":`{"model":"claude"}`,
		"pi/models.json":  `["claude","gpt4"]`,
	})
	return nil
}
```
