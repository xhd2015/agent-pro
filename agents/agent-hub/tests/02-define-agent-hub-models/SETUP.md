## Preconditions
- The Agent Hub model package is available at `github.com/xhd2015/agent-pro/agents/agent-hub/model`.

## Steps
1. Each leaf selects one model scenario.
2. The root runner executes that scenario against the model package.
3. The leaf assertion verifies the result.

```go
import (
    "encoding/json"
    "strings"
    "testing"
    "time"

    "github.com/xhd2015/agent-pro/agents/agent-hub/model"
)

type Request struct {
    Case string
}

type Response struct {
    OK bool
    Error string
    JSON string
    EventType string
}

func Setup(t *testing.T, req *Request) error {
    req.Case = "unset"
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}
    switch req.Case {
    case "normalized-event-valid-minimal":
        evt := model.NormalizedEvent{EventType: model.EventSessionStarted, Runner: "codex"}
        err := evt.Validate()
        resp.OK = err == nil
        if err != nil { resp.Error = err.Error() }
    case "normalized-event-missing-event-type":
        err := (model.NormalizedEvent{Runner: "codex"}).Validate()
        if err != nil { resp.Error = err.Error() }
    case "normalized-event-missing-runner":
        err := (model.NormalizedEvent{EventType: model.EventSessionStarted}).Validate()
        if err != nil { resp.Error = err.Error() }
    case "normalized-event-unknown-event-type":
        err := (model.NormalizedEvent{EventType: "agent.unknown", Runner: "codex"}).Validate()
        if err != nil { resp.Error = err.Error() }
    case "normalized-event-payload-round-trip":
        evt := model.NormalizedEvent{EventType: model.EventSessionStarted, Runner: "codex", Payload: json.RawMessage(`{"nested":{"value":1}}`)}
        data, err := json.Marshal(evt)
        if err != nil { return nil, err }
        var parsed model.NormalizedEvent
        if err := json.Unmarshal(data, &parsed); err != nil { return nil, err }
        resp.JSON = string(parsed.Payload)
        resp.OK = parsed.Validate() == nil
    case "envelope-valid-round-trip":
        env := model.Envelope{SchemaVersion: model.SchemaVersionEventV1, EventID: "evt_1", Partition: "2026-06-10", Offset: 0, ReceivedAt: time.Date(2026, 6, 10, 1, 2, 3, 0, time.UTC), Event: model.NormalizedEvent{EventType: model.EventSessionStarted, Runner: "codex"}}
        data, err := json.Marshal(env)
        if err != nil { return nil, err }
        var parsed model.Envelope
        if err := json.Unmarshal(data, &parsed); err != nil { return nil, err }
        err = parsed.Validate()
        resp.OK = err == nil
        if err != nil { resp.Error = err.Error() }
        resp.JSON = string(data)
    case "envelope-missing-partition":
        err := (model.Envelope{SchemaVersion: model.SchemaVersionEventV1, EventID: "evt_1", Offset: 0, ReceivedAt: time.Now(), Event: model.NormalizedEvent{EventType: model.EventSessionStarted, Runner: "codex"}}).Validate()
        if err != nil { resp.Error = err.Error() }
    case "envelope-negative-offset":
        err := (model.Envelope{SchemaVersion: model.SchemaVersionEventV1, EventID: "evt_1", Partition: "2026-06-10", Offset: -1, ReceivedAt: time.Now(), Event: model.NormalizedEvent{EventType: model.EventSessionStarted, Runner: "codex"}}).Validate()
        if err != nil { resp.Error = err.Error() }
    case "cursor-valid-round-trip":
        cur := model.Cursor{Partition: "2026-06-10", Offset: 42}
        data, err := json.Marshal(cur)
        if err != nil { return nil, err }
        var parsed model.Cursor
        if err := json.Unmarshal(data, &parsed); err != nil { return nil, err }
        err = parsed.Validate()
        resp.OK = err == nil && parsed.Offset == 42
        if err != nil { resp.Error = err.Error() }
    case "cursor-bad-partition-format":
        err := (model.Cursor{Partition: "2026/06/10", Offset: 0}).Validate()
        if err != nil { resp.Error = err.Error() }
    case "fetch-response-empty-batch":
        fr := model.FetchResponse{ConsumerID: "dashboard", Events: []model.Envelope{}, PreviousCursor: model.Cursor{Partition: "2026-06-10", Offset: 2}, NextCursor: model.Cursor{Partition: "2026-06-10", Offset: 2}}
        data, err := json.Marshal(fr)
        if err != nil { return nil, err }
        resp.JSON = string(data)
        resp.OK = strings.Contains(resp.JSON, `"events":[]`)
    case "mock-config-hooks-round-trip":
        cfg := model.MockConfig{Version: "agent-pro.fake-runner.v1", Runner: "fake-codex", Hooks: []model.MockHook{{At: "before_stdout", Event: "SessionStart", Payload: json.RawMessage(`{"ok":true}`)}}}
        data, err := json.Marshal(cfg)
        if err != nil { return nil, err }
        var parsed model.MockConfig
        if err := json.Unmarshal(data, &parsed); err != nil { return nil, err }
        resp.OK = len(parsed.Hooks) == 1 && parsed.Hooks[0].Event == "SessionStart"
        resp.JSON = string(data)
    default:
        resp.Error = "unknown test case: " + req.Case
    }
    return resp, nil
}
```
