---
name: llm-mock
description: >-
  OpenAI-compatible HTTP mock server for testing agents without real LLM calls.
  Loads a JSON config of request→response exchanges and returns pre-configured
  responses. Use when writing doctests or integration tests for agent runners.
---

You are an expert at using llm-mock to test agent and LLM integrations without real API calls.

llm-mock is an OpenAI-compatible HTTP server that:
- Serves POST /v1/chat/completions (streaming and non-streaming)
- Serves POST /v1/responses (OpenAI Responses API)
- Serves GET /v1/models
- Serves GET /admin/requests (records all received requests)
- Loads config from --config flag or LLM_MOCK_CONFIG env var
- Supports port fallback (tries port, port+1, ... port+99)

Config format (JSON):

```json
{
  "port": 8080,
  "exchanges": [
    {
      "request": { "role": "user", "content": "capital of France", "index": 0 },
      "response": { "content": "Paris", "finish_reason": "stop" }
    }
  ]
}
```

Matching rules:
- `index` >= 0: match when request counter equals index
- `index` = -1 (or omitted): sequential match (first request matches first exchange, etc.)
- `role`: exact match; empty role matches any
- `content`: substring match; empty content matches any

Usage:

```sh
go build -o llm-mock ./agent/llm/llm-mock
llm-mock --config config.json
# Server prints listening address to stdout (e.g. :8080)
```

Point your agent runner at the mock:

```sh
OPENAI_BASE_URL=http://localhost:8080/v1 opencode ...
```

For testing, use GET /admin/requests to verify what the client sent.