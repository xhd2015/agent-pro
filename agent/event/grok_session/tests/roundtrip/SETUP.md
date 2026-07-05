# Scenario

**Feature**: semantic roundtrip compatibility across wire → events → wire → events

```
# first pass: grok wire → canonical events
wire₁ -> FromUpdatesJSONL -> events₁

# second pass: canonical events → wire → canonical events again
events₁ -> ToSession -> ToWireLines -> wire₂ -> FromUpdatesJSONL -> events₂

# assert semantic equality (not byte-exact wire)
SemanticEqual(events₁, events₂)
```

## Preconditions

- Roundtrip leaves use `req.Target = "roundtrip"`.
- Assertions compare `resp.Events1` and `resp.Events2` via `SemanticEqual`.
- Byte-exact wire reproduction is out of scope; chunk boundaries may differ.

## Steps

1. Set `req.Target = "roundtrip"`.
2. Leaf SETUPs seed `req.WireLines` with representative session wire input.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	return nil
}
```