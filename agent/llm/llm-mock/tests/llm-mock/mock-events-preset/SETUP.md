# Scenario

**Feature**: `--mock-events-preset=<name>` seeds global FIFO `genQueue` after config exchanges

```
# optional catalog (no server)
llm-mock --mock-events-preset=list -> stdout catalog, exit 0

# server startup loads preset into genQueue
config loader -> genQueue append (preset AgentEvents)
HTTP request -> config exchanges[] (prefix) -> genQueue dequeue -> genStream random fallback
```

## Preconditions

- `--mock-events-preset` is a standalone server flag (`go flag` in `main`).
- Preset names (kebab-case): `simple`, `think-message`, `multi-think`, `tool-bash`, `tool-read`, `think-tool-message`.
- `list` prints catalog to stdout and exits 0 — no listener, no HTTP traffic.
- Unknown preset name errors before the server listens.
- Preset queue is global FIFO: each HTTP serve dequeues one `AgentEvent` (Option A — side requests consume queue).
- After preset queue drained, existing `genStream` random fallback applies (not `no_match`).
- Preset merges with config `exchanges[]`; does not replace prefix exchanges.

## Steps

1. Grouping `Setup` documents preset pipeline and `Request.MockEventsPreset` / `Request.CatalogOnly`.
2. Leaves narrow preset name, config prefix depth, and HTTP request count.
3. `Run` passes `--mock-events-preset` on server start, or runs catalog-only exec when `CatalogOnly` is set.
4. Leaf `Assert` checks stdout catalog, startup errors, HTTP bodies, and/or `--agent-events-file` JSONL order.

## Context

- `Request.MockEventsPreset` — preset name or `list` for catalog leaves.
- `Request.CatalogOnly` — when true, `Run` execs `llm-mock --mock-events-preset=<value>` without starting HTTP server.
- `Response.AgentEventsLines` — served `AgentEvent` JSONL from `--agent-events-file` (think vs message distinction).
- MVP catalog preset event shapes are fixed in `mockpreset` package (implementer).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Endpoint = "/v1/chat/completions"
	req.Method = "POST"
	return nil
}
```