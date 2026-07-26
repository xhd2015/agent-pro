# AgentEvent Tool Input Print Details

# DSN (Domain Specific Notion)

The compact trace printer receives canonical `AgentEvent` JSONL lines from
agent-hub and knowledge-hub sessions. A `tool_call` event has a tool name and a
structured `tool_input` map. The printer renders that event as a compact block:
one header identifying the tool and one or more detail lines explaining what the
tool was asked to do.

For lookup-style tools such as glob/search, the useful detail is usually the
pattern or query from `tool_input`. A bare `SEARCH` header is insufficient for
debugging a maintain-topic trace.

## Version

0.0.2

## Decision Tree

```
canonical AgentEvent tool_call line
├── tool=glob
│   └── input has pattern              -> SEARCH block includes pattern
├── tool=skill
│   └── input has name                 -> SKILL block includes skill name
├── tool=webfetch
│   └── input has url                  -> WEBFETCH block includes URL
├── tool=todowrite
│   └── input has todos[]              -> TODO block includes todo details
└── tool=skynet-base_get_doc_content
    └── input has doc                  -> tool block includes search/doc URL
```

## Test Leaves

| Leaf | Description |
|---|---|
| `glob-pattern-details` | `glob` tool input pattern is visible below the `SEARCH` header. |
| `skill-name-details` | `skill` tool input name is visible below the `SKILL` header. |
| `webfetch-url-details` | `webfetch` tool input URL is visible below the `WEBFETCH` header. |
| `todowrite-todos-details` | `todowrite` todos are visible below the `TODO` header. |
| `skynet-doc-search-details` | Confluence search/doc URL is visible below the skynet tool header. |

## How to Run

```sh
doctest vet ./agent/event/print/tests/agentevent_tool_inputs
doctest test ./agent/event/print/tests/agentevent_tool_inputs/...
```

```go
import (

	"testing"

	"github.com/xhd2015/agent-pro/agent/event/print"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Line string
}

type Response struct {
	Output string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return &Response{Output: print.FormatTraceLine(req.Line)}, nil
}
```
