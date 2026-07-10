# Refactor Plan: Add github.com/slack-go/slack and Refactor debug/slack-send

**Context**  
This plan addresses the request to add the well-known community Slack Go library (`github.com/slack-go/slack`) and refactor the existing debug script.

The current implementation (`script/debug/slack-send/main.go`) uses only the Go standard library (`net/http`, `encoding/json`, etc.) to:
- Load configuration from `slack-config.json` (discovered by walking up to `go.mod`).
- Resolve channel names/IDs using a simple map.
- Send a message via direct POST to `https://slack.com/api/chat.postMessage` using the `botToken`.

The goal is to switch the sending logic to the SDK while keeping **user-facing behavior, CLI interface, output format, error messages, and config file format 100% identical**.

This plan was derived from the TDD requirement document `REQUIREMENT-DESIGN-refactor-slack-send-to-slack-go-sdk.md`.

**Status**  
- This is a clarification + design artifact.  
- Full TDD flow (designer → RED doctests → seal → implementer) is in progress per `/doctest-tdd`.  
- No source files were modified by the orchestrator.

---

## 1. Data Models & Storage Layout

### Storage (unchanged)
- `slack-config.json` (at project root or discovered via upward search for `go.mod`).
- Contains:
  - `botToken` (primary for sending)
  - `appToken`
  - `knownChannels` (map[string]string, e.g. `"#general": "C0ALE44K5J6"`)
  - `defaultChannelId`, `defaultChannelName`
  - Preserved original fields: `source`, `extractedAt`, `config`, `plugins`

### In-Code Models
- `SlackConfig` struct (kept nearly identical):
  ```go
  type SlackConfig struct {
      Source             string            `json:"source"`
      BotToken           string            `json:"botToken"`
      AppToken           string            `json:"appToken"`
      DefaultChannelId   string            `json:"defaultChannelId"`
      DefaultChannelName string            `json:"defaultChannelName"`
      KnownChannels      map[string]string `json:"knownChannels"`
      Config             json.RawMessage   `json:"config"`
  }
  ```
- New (SDK):
  - `*slack.Client` created via `slack.New(token)`
  - `slack.MsgOptionText(text, false)`
  - `PostMessage(...)` returns `(channel string, timestamp string, err error)`

No new files, databases, or persistent state.

---

## 2. CLI Contract (must be preserved exactly)

**Invocation examples** (unchanged behavior):
```bash
go run ./script/debug/slack-send
go run ./script/debug/slack-send "#general"
go run ./script/debug/slack-send "#general" "Hello from Go debug script"
go run ./script/debug/slack-send C0ALE44K5J6 "custom message here"
go run ./script/debug/slack-send -h
```

**Success output** (must match exactly, including trailing newline):
```
Sending to channel=C0ALE44K5J6: "Hello slack"
Using config from: /path/to/slack-config.json
OK ts=1783398010.628649 channel=C0ALE44K5J6
```

**Error outputs** (preserve current messages where possible):
- Config problems → `failed to load config ...`
- Missing token → `botToken is empty in ...`
- Send failure → `send failed: ...` or `slack error: ...`

Help text must remain identical.

---

## 3. Full Refactor Steps (for Implementer)

### Step 1: Add Dependency
- Run (or simulate in TDD):
  ```bash
  go get github.com/slack-go/slack
  go mod tidy
  ```
- Expected `go.mod` change:
  ```go
  require github.com/slack-go/slack v0.27.0
  ```
- `github.com/gorilla/websocket` will appear as indirect (already a direct dep in the project — no conflict).
- Test-only deps (`testify`, `go-test/deep`) from the library's `go.mod` **do not** appear as direct requires in the consumer `go.mod` when only using the Web API client.

### Step 2: Refactor `script/debug/slack-send/main.go`
Keep the file as `package main`. Make **minimal** changes.

**Keep unchanged (byte-for-byte where possible):**
- `const defaultConfigPath`
- `type SlackConfig`
- `func main()` structure for arg parsing, config loading, channel resolution, printing
- `func printHelp()`
- `func findConfigPath()`
- `func loadConfig()`
- `func resolveChannel()`

**Changes:**
1. Add import:
   ```go
   import "github.com/slack-go/slack"
   ```
2. Remove:
   - `type slackResponse struct { ... }`
   - `func sendSlackMessage(...)` (the entire raw HTTP implementation using `bytes`, `net/http`, `io`)
3. Replace the send block (after the "Sending..." and "Using config..." prints) with:
   ```go
   api := slack.New(cfg.BotToken)
   // channelID already resolved by resolveChannel()
   _, ts, err := api.PostMessage(channelID, slack.MsgOptionText(text, false))
   if err != nil {
       fmt.Fprintf(os.Stderr, "send failed: %v\n", err)
       os.Exit(1)
   }
   fmt.Printf("OK ts=%s channel=%s\n", ts, channelID)
   ```
4. Remove unused imports (`bytes`, `net/http`, `io`) after the change.
5. The SDK `PostMessage` returns `(string, string, error)` — use the second return value directly as the timestamp (matches previous `resp.TS`).

**Optional (but recommended for future):**
- Keep a thin wrapper if it helps testability, but do **not** change observable behavior.

### Step 3: Config & Discovery
- `slack-config.json` is **never modified**.
- `botToken` is passed directly to `slack.New(botToken)`.
- All discovery logic (`findConfigPath`) remains identical.

### Step 4: Output & Error Handling
- Preserve every `fmt.Printf` / `fmt.Fprintf` line and exit codes exactly.
- SDK errors will be reported under the existing `"send failed: %v"` prefix.
- Final success line must still end with a newline.

### Step 5: Build / Run / Verification
- After changes:
  ```bash
  go run ./script/debug/slack-send/main.go
  go run ./script/debug/slack-send/main.go "#general" "test message"
  go list -m github.com/slack-go/slack
  go mod tidy
  ```
- The script must still work from any subdirectory (config walk logic preserved).
- Real message must appear in the target Slack channel when using the live token.

### Step 6: Testing Strategy (driven by doctests)
- Pure functions (`resolveChannel`, `loadConfig`, `findConfigPath`, arg parsing, help) remain easily unit-testable.
- Send path becomes an SDK call.
- Doctests will:
  - Assert exact stdout/stderr for all scenarios.
  - Use `slacktest` (from the library) for isolated send tests where possible.
  - Use real Slack calls under `integration` / `slow` labels.
  - Verify `go run` behavior and module state.

---

## 4. Non-Goals / Out of Scope
- Do not change `slack-config.json` content or location.
- Do not add new CLI flags, blocks, threads, file uploads, or other features.
- Do not move the script out of `script/debug/`.
- Do not touch any other code in the repository.
- Do not introduce major new abstractions unless strictly required for the doctests.
- Keep changes as small as possible while switching to the SDK.

---

## 5. Risks & Mitigations
| Risk | Mitigation |
|------|------------|
| Slight difference in SDK error messages | Assert on prefixes ("send failed:") in tests; accept small wording improvements with justification |
| Real Slack calls create messages | Acceptable for a debug script; use labels to skip in fast runs |
| Library has "no major version" note | Pin via `go get` / go.mod; library is mature (v0.27.0) and actively released |
| Adding dep to root go.mod | Acceptable for debug tooling; impact is low (reuses existing websocket dep) |

---

## 6. Verification Checklist (post-implementation)
- [ ] `doctest vet ./tests/<feature>`
- [ ] `doctest test ./tests/<feature>` → all GREEN
- [ ] `go run ./script/debug/slack-send/main.go` produces expected output + real Slack message
- [ ] `go list -m github.com/slack-go/slack` succeeds
- [ ] `go mod tidy` is clean
- [ ] `doctest test ./...` shows no regressions
- [ ] `git diff ./tests/<feature>` is clean (after seal)
- [ ] Identical behavior to pre-refactor version for all documented cases

---

## 7. Files Expected to Change
- `go.mod` / `go.sum` (dependency)
- `script/debug/slack-send/main.go` (refactor send logic only)

**No changes to:**
- `slack-config.json`
- Any other source files

---

**This plan is designed to be driven by sealed doctests.**

The corresponding TDD requirement lives at:
`REQUIREMENT-DESIGN-refactor-slack-send-to-slack-go-sdk.md`

Once the designer subagent produces the doctest tree and it is sealed, the implementer will follow the steps above to make the tests pass.