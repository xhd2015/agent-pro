## Preconditions
- Regular config files (no API keys, no auth tokens) get 0644 permission.
- The operation has been set to "import".

## Steps
1. Create a zip with non-sensitive files only.
2. Run import.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "import"
	req.Agent = "pi"
	createZip(t, req.ZipPath, map[string]string{
		"pi/settings.json":  `{"model":"claude"}`,
		"pi/models.json":    `["claude","gpt4"]`,
		"opencode/skills/deploy/SKILL.md": "# Deploy\n",
		"crush/data/crush.json": `{"projects":[]}`,
	})
	return nil
}
```
