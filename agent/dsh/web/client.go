// Package web controls sessions on a running DSH Web instance. Unlike the
// one-shot CLI handoff, callers can persist a session ID before submitting input
// and reuse it for follow-ups and browser-launch retries.
package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/open"
)

const Runner = "dsh-web"

// Listen contains discovery credentials. Never persist or log Token in a
// consumer's session binding; read the current record on each operation.
type Listen struct {
	Origin string `json:"origin"`
	Token  string `json:"token,omitempty"`
}

// Client has injectable transport/discovery/opener boundaries for in-process tests.
// Its zero value uses DSH_HOME (or ~/.dsh), a bounded HTTP client, and the OS opener.
type Client struct {
	Home        string
	ReadListen  func() (Listen, error)
	HTTPClient  *http.Client
	OpenBrowser func(rawURL, browser string) error
}

func NormalizeBrowser(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "", "default":
		return "", nil
	case "brave", "chrome", "firefox", "edge", "safari", "opera":
		return s, nil
	default:
		return "", fmt.Errorf("unknown DSH browser: %s", raw)
	}
}

func (c Client) listen() (Listen, error) {
	var rec Listen
	if c.ReadListen != nil {
		var err error
		rec, err = c.ReadListen()
		if err != nil {
			return Listen{}, err
		}
	} else {
		home := c.Home
		if home == "" {
			home = os.Getenv("DSH_HOME")
		}
		if home == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return Listen{}, err
			}
			home = filepath.Join(userHome, ".dsh")
		}
		data, err := os.ReadFile(filepath.Join(home, "web-listen.json"))
		if err != nil {
			return Listen{}, fmt.Errorf("DSH web is not serving; start it with: dsh web")
		}
		if err := json.Unmarshal(data, &rec); err != nil {
			return Listen{}, fmt.Errorf("invalid DSH web listen record; restart with: dsh web")
		}
	}
	u, err := url.Parse(rec.Origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return Listen{}, fmt.Errorf("DSH web has no valid listen origin; start it with: dsh web")
	}
	return rec, nil
}

// Create registers a workspace and an empty session under a caller-persisted
// identity. A retry can reuse that blank session. No prompt is submitted and no
// browser is opened: the caller must durably save the returned ID first.
func (c Client) Create(ctx context.Context, directory, sessionID string) (string, error) {
	created, err := c.CreateWithOptions(ctx, CreateOptions{Workspace: directory, SessionID: sessionID})
	return created.SessionID, err
}

// CreateWithOptions initializes a reserved session and verifies the Host accepted
// an explicit model before any caller can submit input or open the browser.
func (c Client) CreateWithOptions(ctx context.Context, opts CreateOptions) (CreatedSession, error) {
	sessionID := opts.SessionID
	selected, err := normalizeModelSelection(opts.ModelSelection)
	if err != nil {
		return CreatedSession{}, err
	}
	if strings.TrimSpace(sessionID) == "" {
		return CreatedSession{}, fmt.Errorf("DSH creation requires a stable session ID")
	}
	directory, err := filepath.Abs(opts.Workspace)
	if err != nil {
		return CreatedSession{}, err
	}
	st, err := os.Stat(directory)
	if err != nil {
		return CreatedSession{}, err
	}
	if !st.IsDir() {
		return CreatedSession{}, fmt.Errorf("DSH workspace is not a directory")
	}
	var workspace struct {
		Workspace struct {
			ID string `json:"workspaceId"`
		} `json:"workspace"`
	}
	if err := c.rpc(ctx, "workspace/create", map[string]any{"request": map[string]string{"path": directory}}, &workspace); err != nil {
		return CreatedSession{}, err
	}
	if workspace.Workspace.ID == "" {
		return CreatedSession{}, fmt.Errorf("DSH workspace/create returned no workspace ID")
	}
	request := map[string]any{"workspaceId": workspace.Workspace.ID, "sessionId": sessionID}
	if selected != nil {
		request["modelSelection"] = selected
	}
	var session CreatedSession
	if err := c.rpc(ctx, "session/create", map[string]any{"request": request}, &session); err != nil {
		return CreatedSession{}, err
	}
	if session.SessionID == "" {
		return CreatedSession{}, fmt.Errorf("DSH session/create returned no session ID")
	}
	if session.SessionID != sessionID {
		return CreatedSession{}, fmt.Errorf("DSH could not reuse the reserved session identity; refusing to replace its conversation")
	}
	if selected != nil {
		accepted, err := normalizeModelSelection(session.ModelSelection)
		if err != nil || accepted == nil || accepted.Provider != selected.Provider || accepted.Model != selected.Model || (selected.ReasoningEffort != "" && accepted.ReasoningEffort != selected.ReasoningEffort) {
			return CreatedSession{}, fmt.Errorf("DSH Web did not confirm the requested Session-only model selection; update DSH before retrying (no prompt was sent)")
		}
		session.ModelSelection = accepted
	}
	return session, nil
}

// Send queues a prompt (or stores a draft). A caller-supplied requestID is reused
// across retries so an ambiguous transport failure does not enqueue twice.
func (c Client) Send(ctx context.Context, sessionID, requestID, text string, noSubmit bool) error {
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(text) == "" || strings.TrimSpace(requestID) == "" {
		return fmt.Errorf("DSH session, request ID and prompt are required")
	}
	request := map[string]any{"sessionId": sessionID}
	method := "session/prompt"
	if noSubmit {
		method = "session/setDraft"
		request["text"] = text
	} else {
		request["requestId"] = requestID
		request["mode"] = "queue"
		request["content"] = []map[string]string{{"type": "text", "text": text}}
	}
	return c.rpc(ctx, method, map[string]any{"request": request}, nil)
}

// Open opens the existing session. Opener errors are deliberately redacted:
// platform diagnostics can contain the authenticated URL.
func (c Client) Open(sessionID, browser string) error {
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("DSH session ID is required")
	}
	browser, err := NormalizeBrowser(browser)
	if err != nil {
		return err
	}
	rec, err := c.listen()
	if err != nil {
		return err
	}
	u, _ := url.Parse(rec.Origin)
	u.Path, u.RawPath, u.Fragment = "/", "", ""
	q := url.Values{"session": {sessionID}}
	if rec.Token != "" {
		q.Set("token", rec.Token)
	}
	u.RawQuery = q.Encode()
	opener := c.OpenBrowser
	if opener == nil {
		opener = func(raw, app string) error { _, err := open.Browser(raw, app); return err }
	}
	if err := opener(u.String(), browser); err != nil {
		return fmt.Errorf("DSH session %s exists, but the browser could not be opened; retry opening the same session", sessionID)
	}
	return nil
}

func (c Client) rpc(ctx context.Context, method string, args any, result any) error {
	rec, err := c.listen()
	if err != nil {
		return err
	}
	rpcID := uuid.NewString()
	body, err := json.Marshal(map[string]any{"type": "client-request", "rpcId": rpcID, "method": method, "payload": map[string]any{"args": args}})
	if err != nil {
		return err
	}
	u, _ := url.Parse(rec.Origin)
	u.Path, u.RawPath, u.RawQuery, u.Fragment = "/api/"+method, "", "", ""
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if rec.Token != "" {
		req.Header.Set("Authorization", "Bearer "+rec.Token)
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("DSH web request failed; ensure dsh web is running (delivery may have been accepted; retry with the same request ID)")
	}
	defer res.Body.Close()
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return fmt.Errorf("DSH web rejected its recorded token; restart with: dsh web")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("DSH RPC %s failed: HTTP %d", method, res.StatusCode)
	}
	var envelope struct {
		Type   string `json:"type"`
		RPCID  string `json:"rpcId"`
		Result *struct {
			OK    bool            `json:"ok"`
			Value json.RawMessage `json:"value"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		} `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&envelope); err != nil {
		return fmt.Errorf("DSH RPC %s returned invalid JSON", method)
	}
	if envelope.Type != "server-response" || envelope.RPCID != rpcID || envelope.Result == nil {
		return fmt.Errorf("DSH RPC %s returned an invalid envelope", method)
	}
	if !envelope.Result.OK {
		// Do not echo arbitrary server text, which may carry credentials.
		if envelope.Result.Error != nil {
			return fmt.Errorf("DSH RPC %s failed (%s)", method, envelope.Result.Error.Code)
		}
		return fmt.Errorf("DSH RPC %s failed", method)
	}
	if result != nil {
		if err := json.Unmarshal(envelope.Result.Value, result); err != nil {
			return fmt.Errorf("DSH RPC %s returned an invalid result", method)
		}
	}
	return nil
}
