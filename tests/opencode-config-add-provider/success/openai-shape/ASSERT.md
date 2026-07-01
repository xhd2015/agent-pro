## Expected

- Exit code 0.
- `provider.oai.npm` == `@ai-sdk/openai-compatible`.
- `provider.oai.models.gpt-4o.name` == `gpt-4o`.

## Side Effects

- Global config file created with the openai-shape provider entry.

## Errors

- None.

## Exit Code

- 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessCommon(t, req, resp, "oai")
	entry, _, e := readProviderEntry(resp.ConfigPath, "oai")
	if e != nil {
		t.Fatal(e)
	}
	if got := entry["npm"]; got != "@ai-sdk/openai-compatible" {
		t.Fatalf("npm = %v, want @ai-sdk/openai-compatible", got)
	}
}
```
