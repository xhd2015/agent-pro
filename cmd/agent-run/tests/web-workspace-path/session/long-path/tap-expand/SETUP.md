# Scenario

**Feature**: session page long workspace collapses by default and expands on tap

```
# same WorkspacePath component on session surface
seed session(workspace=deep) -> /sessions/<id>
  -> collapsed: short …/ label
  -> tap workspace-toggle -> full path in workspace-label
```

## Preconditions

- Shared component required by option A on session header (expect RED:
  session currently renders raw full path without toggle/shorten).
- Flat session id route.

## Steps

1. Seed flat session with long `workspace`.
2. Start web with explicit token; open `/sessions/<id>`.
3. Assert collapsed short label, then expand to full path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "session-long-tap-expand"

	if err := seedFlatSessionWithWorkspace(t, req.Home, req.SessionID, "fake-codex", "idle", req.WorkspacePath); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	full := jsString(req.WorkspacePath)
	expectedShort := jsString(shortWorkspaceLabel(req.WorkspacePath))
	sessionPath := "/sessions/" + req.SessionID

	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + jsWaitWorkspaceVisible() + jsWorkspaceLabelText() + `
const before = await workspaceLabelText();
if (before !== '` + expectedShort + `') {
  throw new Error('expected collapsed short session workspace ` + expectedShort + `, got: ' + before);
}
const toggle = page.locator('[data-testid="workspace-toggle"]');
await toggle.waitFor({ state: 'visible', timeout: 15000 });
await toggle.click();
await page.waitForTimeout(100);
const after = await workspaceLabelText();
if (after !== '` + full + `') {
  throw new Error('expected expanded full path on session, got: ' + after);
}
const expanded = await toggle.getAttribute('aria-expanded');
if (expanded !== 'true') {
  throw new Error('expected aria-expanded=true, got: ' + expanded);
}
`

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
