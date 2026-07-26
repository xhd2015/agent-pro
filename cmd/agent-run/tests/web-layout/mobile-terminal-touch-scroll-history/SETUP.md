# Scenario

**Feature**: mobile touch pan on terminal modal reveals older LINE_* scrollback

```
seed finished codex-tty session + fake PTY with many LINE_xxx rows
  → open /sessions/<id> → Terminal button → terminal-surface
  → synthetic vertical touch pan (finger down)
  → min visible LINE_* index decreases (older history)
```

## Preconditions

- `playwright-debug` on PATH.
- Flat session store + `codex-tty-registry` entry pointing at fake PTY websocket.
- Fake PTY streams ≥120 `LINE_NNN` rows so mobile viewport cannot show all.
- Open API.
- Label: `ui-automation`.

## Steps

1. Start fake ptywrap that streams LINE_000… scrollback on WS connect.
2. Seed flat session with `terminal_session_id` + registry listen_addr.
3. Start agent-run web; open session; open Terminal modal; touch-pan; assert older lines.

```go
import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
)

const layoutTermTouchSessionID = "web_term_touch_scroll"
const layoutTermTouchTerminalID = "session-touch-1"
const layoutTermTouchScrollbackLines = 120

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "terminal-touch-scroll-history"
	req.WebTokenMode = "omit"
	runner := "codex-tty"

	ptyAddr := startLayoutScrollbackPty(t, layoutTermTouchTerminalID, layoutTermTouchScrollbackLines)
	if err := seedLayoutTerminalTouchSession(t, req.Home, runner, layoutTermTouchSessionID, layoutTermTouchTerminalID, ptyAddr); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	// Sanity: terminal API available before browser.
	if err := waitLayoutTerminalAvailable(req.BaseURL, layoutTermTouchSessionID, "", 15*time.Second); err != nil {
		return err
	}

	sessionPath := "/sessions/" + layoutTermTouchSessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle', timeout: 45000 });
await page.waitForSelector('[data-testid="chat-active"], [data-testid="message-list"]', { timeout: 20000 });
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 20000 });
await terminalButton.click();
const surface = page.locator('[data-testid="terminal-surface"]');
await surface.waitFor({ state: 'visible', timeout: 20000 });
await page.locator('[data-testid="terminal-surface"] .xterm').waitFor({ state: 'visible', timeout: 20000 });
await page.waitForFunction(() => {
  const el = document.querySelector('[data-testid="terminal-surface"] .xterm-rows');
  const t = (el && el.textContent) || '';
  return /LINE_\d{3}/.test(t);
}, null, { timeout: 20000 });
await page.waitForTimeout(400);

function parseLineMarkers(text) {
  const re = /LINE_(\d{3})/g;
  const nums = [];
  let m;
  while ((m = re.exec(text || '')) !== null) nums.push(parseInt(m[1], 10));
  if (nums.length === 0) return { count: 0, min: null, max: null };
  return { count: nums.length, min: Math.min(...nums), max: Math.max(...nums) };
}

async function readTerminalSnapshot() {
  return page.evaluate(() => {
    const root = document.querySelector('[data-testid="terminal-surface"]');
    if (!root) return { ok: false, text: '' };
    const rows = root.querySelector('.xterm-rows');
    const text = (rows && rows.textContent) || root.textContent || '';
    let scrollTop = null;
    let bestRange = 0;
    for (const el of root.querySelectorAll('*')) {
      const sh = el.scrollHeight || 0;
      const ch = el.clientHeight || 0;
      if (sh > ch + 10 && sh - ch > bestRange) {
        bestRange = sh - ch;
        scrollTop = el.scrollTop;
      }
    }
    return { ok: true, text: text.slice(0, 4000), scrollTop };
  });
}

async function touchPanVertical(distancePx, steps) {
  return page.evaluate(({ distancePx, steps }) => {
    const root = document.querySelector('[data-testid="terminal-surface"]');
    if (!root) throw new Error('no terminal-surface for touch pan');
    const rect = root.getBoundingClientRect();
    const x = rect.left + rect.width / 2;
    const startY = rect.top + rect.height * 0.35;
    const endY = startY + distancePx;
    const fire = (type, y, touchesList) => {
      const touch = new Touch({
        identifier: 1,
        target: root,
        clientX: x,
        clientY: y,
        pageX: x,
        pageY: y,
        screenX: x,
        screenY: y,
        radiusX: 2,
        radiusY: 2,
        rotationAngle: 0,
        force: 1,
      });
      root.dispatchEvent(new TouchEvent(type, {
        bubbles: true,
        cancelable: true,
        composed: true,
        touches: touchesList(touch),
        targetTouches: touchesList(touch),
        changedTouches: [touch],
      }));
    };
    fire('touchstart', startY, (t) => [t]);
    for (let i = 1; i <= steps; i++) {
      const y = startY + ((endY - startY) * i) / steps;
      fire('touchmove', y, (t) => [t]);
    }
    fire('touchend', endY, () => []);
    return { startY, endY, distancePx, steps };
  }, { distancePx, steps });
}

const beforeSnap = await readTerminalSnapshot();
const beforeLines = parseLineMarkers(beforeSnap.text || '');
if (beforeLines.count === 0) {
  throw new Error('no LINE_xxx markers visible after terminal open');
}
await touchPanVertical(320, 16);
await page.waitForTimeout(200);
await touchPanVertical(320, 16);
await page.waitForTimeout(250);
const afterSnap = await readTerminalSnapshot();
const afterLines = parseLineMarkers(afterSnap.text || '');
let touchScrolled = false;
if (beforeLines.min != null && afterLines.min != null && afterLines.min < beforeLines.min) {
  touchScrolled = true;
} else if (
  beforeSnap.scrollTop != null &&
  afterSnap.scrollTop != null &&
  afterSnap.scrollTop < beforeSnap.scrollTop - 5
) {
  touchScrolled = true;
}
if (!touchScrolled) {
  throw new Error(
    'touch pan did not reveal older scrollback: beforeMin=' + beforeLines.min +
    ' afterMin=' + afterLines.min +
    ' scrollTop before=' + beforeSnap.scrollTop + ' after=' + afterSnap.scrollTop,
  );
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}

func startLayoutScrollbackPty(t *testing.T, termID string, lines int) string {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		requestedID := r.URL.Query().Get("session_id")
		if requestedID == "" {
			requestedID = termID
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"`+requestedID+`"}`))
		var b strings.Builder
		for i := 0; i < lines; i++ {
			fmt.Fprintf(&b, "LINE_%03d scrollback filler row for mobile touch pan testing\r\n", i)
		}
		_ = conn.WriteMessage(websocket.BinaryMessage, []byte(b.String()))
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.BinaryMessage {
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("echo:"+string(msg)))
			}
		}
	})
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"` + termID + `","status":"running"}]`))
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake pty: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		_ = srv.Close()
		_ = ln.Close()
	})
	return ln.Addr().String()
}

func seedLayoutTerminalTouchSession(t *testing.T, home, runner, chatID, termID, listenAddr string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", chatID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(home, "sessions", ".layout"), []byte(`{"version":2}`+"\n"), 0644); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":              runner,
		"session_id":          chatID,
		"terminal_session_id": termID,
		"status":              "finished",
		"created_at":          now,
		"updated_at":          now,
		"workspace":           filepath.Join(home, "workspace"),
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	events := `{"type":"message","role":"user","text":"show terminal scrollback","timestamp":1}` + "\n" +
		`{"type":"message","role":"assistant","text":"open the Terminal button to view PTY","timestamp":2}` + "\n"
	if err := os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644); err != nil {
		return err
	}
	regDir := filepath.Join(home, runner+"-registry")
	if err := os.MkdirAll(regDir, 0755); err != nil {
		return err
	}
	entry := map[string]any{
		"session_id":  termID,
		"listen_addr": listenAddr,
		"pid":         os.Getpid(),
		"created_at":  time.Now().Format(time.RFC3339Nano),
	}
	entryBytes, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(regDir, termID+".json"), entryBytes, 0644)
}

func waitLayoutTerminalAvailable(baseURL, sessionID, bearer string, timeout time.Duration) error {
	client := &http.Client{Timeout: 3 * time.Second}
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/sessions/" + sessionID + "/terminal"
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		httpReq, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		if bearer != "" {
			httpReq.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := client.Do(httpReq)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && strings.Contains(string(body), `"available":true`) {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("terminal not available for %s within %s", sessionID, timeout)
}
```
