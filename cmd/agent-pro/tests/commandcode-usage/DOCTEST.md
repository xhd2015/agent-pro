# Command Code Usage Tests

Doc-style tests for `agent/commandcode`, the read-only Command Code account
client behind `agent-pro commandcode <usage|whoami|credits|...>`. Each leaf
serves a synthetic API over `httptest`, runs the composite usage fetch, and
checks the rendered overlay or the merged JSON payload.

# DSN (Domain Specific Notion)

`agent-pro commandcode usage` mirrors the cmd CLI's usage overlay: it calls
`/alpha/whoami?limits=1`, takes `whoami.org.id` as the org scope, loads
`/alpha/billing/credits` and `/alpha/billing/subscriptions` for that scope in
parallel, then bounds `/alpha/usage/summary` by the subscription's
`currentPeriodStart`. A failing endpoint never aborts the fetch: the field
stays nil and one message per failure lands in `UsageData.Errors`.

Rendering contract (`ProjectUsageView` + `FormatUsage`):
- `USAGE` badge, plan display name, subscription status
- one progress bar measuring against `max(plan allowance, monthly credits) +
  purchased + free`
- `Cycle: $X.XX left · N requests · D days to renewal`
- 5-hour / weekly meters when the API reports window limits
- `Full breakdown at <studio>/<login>/settings/usage`

Overlay layout (one leading space inside the badge, blank lines as shown):

```
 USAGE  Go Plan · active

███████████████████████░░░░░░░ 76% used
Cycle: $2.41 left · 1,040 requests · 5 days to renewal

Usage limits
5-hour  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0%

Weekly  ██░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 8% · resets in 7h 4m

Full breakdown at commandcode.ai/alice/settings/usage
```

## Version

0.0.1

## Decision Tree

```
format?
├── human/          FormatUsage over the projected view, width 80
│   ├── active-plan/      full payload → 76% bar, both meters, Studio URL
│   ├── over-limit/       ~99% usage with saturated 5-hour/weekly windows
│   ├── no-billing/       every billing endpoint 404s → empty state
│   ├── partial-failure/  credits 401 + summary 502 → warnings, degraded body
│   └── forbidden-scope/  summary 403 → the permission message, not an
│                         expired session
└── json/
    └── active-plan/      merged payload with the CLI's key shape
scope?
├── derived-from-whoami/  org and period come from whoami + subscription
└── pinned/               --org / --since overrides win over both
```

Intentional exclusions: credential discovery and auth.json parsing (covered by
package L1 unit tests), and CLI flag parsing (a thin wrapper over the same
package, covered by live smoke runs).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `human/active-plan` | Full payload renders the CLI overlay line for line |
| 2 | `human/over-limit` | Saturated windows render 100% and the bar stays short of full |
| 3 | `human/no-billing` | No billing data → empty state, one warning per failure |
| 4 | `human/partial-failure` | Credits and summary failures degrade instead of aborting |
| 5 | `human/forbidden-scope` | A 403 reports the API's permission message, not `Session expired` |
| 6 | `json/active-plan` | JSON keys are whoami/credits/subscription/summary/errors |
| 7 | `scope/derived-from-whoami` | whoami org scopes credits/subscription/summary |
| 8 | `scope/pinned` | `--org` / `--since` override the derived scope |

## How to Run

```sh
cd external/agent-pro
doctest vet ./cmd/agent-pro/tests/commandcode-usage
doctest test -v ./cmd/agent-pro/tests/commandcode-usage
```

```go
import (
	"context"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/commandcode"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	// Home is the Command Code config dir; root Setup creates it and writes
	// auth.json into it.
	Home string
	// Bodies maps an API path to the raw response body the fake API returns.
	// Paths absent from the map answer 404 with an error envelope.
	Bodies map[string]string
	// Statuses overrides the HTTP status per path (default 200).
	Statuses map[string]int
	// Queries records the raw query string the client sent per path.
	Queries map[string]string
	// AuthHeaders records the Authorization header the client sent per path.
	AuthHeaders map[string]string
	// Format selects the renderer: "human" (default) or "json".
	Format string
	// Width is the terminal width basis passed to FormatUsage.
	Width int
	// Org and Since pin the fetch scope, as --org and --since do.
	Org   string
	Since string
}

type Response struct {
	Data   *commandcode.UsageData
	View   *commandcode.View
	Output string
	JSON   []byte
	Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	server := ServeFixture(t, req.Bodies, req.Statuses, req.Queries, req.AuthHeaders)

	client, err := commandcode.NewClient(req.Home, server.URL)
	if err != nil {
		resp.Err = err
		return resp, nil
	}
	data, err := client.FetchUsageWithOptions(context.Background(), commandcode.UsageOptions{
		OrgID: req.Org,
		Since: req.Since,
	})
	if err != nil {
		resp.Err = err
		return resp, nil
	}
	resp.Data = data
	resp.View = commandcode.ProjectUsageView(data, time.Now())

	if req.Format == "json" {
		raw, err := commandcode.FormatUsageJSON(data)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.JSON = raw
		return resp, nil
	}
	resp.Output = commandcode.FormatUsage(resp.View, commandcode.FormatOptions{Width: req.Width})
	return resp, nil
}
```
