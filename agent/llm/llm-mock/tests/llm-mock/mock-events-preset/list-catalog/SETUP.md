# Scenario

**Feature**: standalone `--mock-events-preset=list` prints catalog and exits without server

```
llm-mock --mock-events-preset=list -> stdout (preset names + descriptions) -> exit 0
(no listener, no HTTP)
```

## Steps

1. Set `CatalogOnly` and `MockEventsPreset=list`.
2. Do not send HTTP requests.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CatalogOnly = true
	req.MockEventsPreset = "list"
	return nil
}
```