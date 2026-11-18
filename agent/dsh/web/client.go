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

// BrowserNone disables automatic session-window opening while retaining normal
// session creation and prompt delivery. It is a persisted browser preference.
const BrowserNone = "none"

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
	case BrowserNone, "brave", "chrome", "firefox", "edge", "safari", "opera":
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
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(text) == "" {
		return fmt.Errorf("DSH session and prompt are required")
	}
	if noSubmit {
		return c.SetDraft(ctx, sessionID, text)
	}
	if strings.TrimSpace(requestID) == "" {
		return fmt.Errorf("DSH request ID is required")
	}
	request := map[string]any{
		"sessionId": sessionID,
		"requestId": requestID,
		"mode":      "queue",
		"content":   []map[string]string{{"type": "text", "text": text}},
	}
	return c.rpc(ctx, "session/prompt", map[string]any{"request": request}, nil)
}

// InjectedMessage is one transcript entry of an injected batch, in authored
// order: a user message, or an assistant message carrying reasoning (thinking)
// or text.
type InjectedMessage struct {
	// Kind selects the transcript role: "user", "thinking" or "assistant".
	Kind string
	// Text is the message body; blank text is rejected before it reaches DSH.
	Text string
}

// InjectResult reports what an injected batch committed.
type InjectResult struct {
	// Seqs are the committed event sequence numbers, in batch order.
	Seqs []int64
	// Started reports whether the agent runs on the batch without another call.
	Started bool
}

// injectableKinds is the closed set DSH accepts, kept here so a typo fails
// before the round trip instead of surfacing as a server-side validation error.
var injectableKinds = map[string]bool{"user": true, "thinking": true, "assistant": true}

// InjectMessages commits one batch of transcript messages in a single DSH
// transaction: the whole batch lands or none of it does, and DSH then lets the
// agent act on it — continuing the running turn, or opening one when idle.
//
// The caller supplies requestID so an ambiguous transport failure can be
// retried without committing the batch twice.
func (c Client) InjectMessages(ctx context.Context, sessionID, requestID string, messages []InjectedMessage) (InjectResult, error) {
	if strings.TrimSpace(sessionID) == "" {
		return InjectResult{}, fmt.Errorf("DSH session is required")
	}
	if strings.TrimSpace(requestID) == "" {
		return InjectResult{}, fmt.Errorf("DSH request ID is required")
	}
	if len(messages) == 0 {
		return InjectResult{}, fmt.Errorf("at least one injected message is required")
	}
	wire := make([]map[string]string, 0, len(messages))
	for _, message := range messages {
		if !injectableKinds[message.Kind] {
			return InjectResult{}, fmt.Errorf("injected message kind must be user, thinking or assistant; got %q", message.Kind)
		}
		if strings.TrimSpace(message.Text) == "" {
			return InjectResult{}, fmt.Errorf("injected %s message has no text", message.Kind)
		}
		wire = append(wire, map[string]string{"kind": message.Kind, "text": message.Text})
	}
	request := map[string]any{"sessionId": sessionID, "requestId": requestID, "messages": wire}
	var value struct {
		Seqs    []int64 `json:"seqs"`
		Started bool    `json:"started"`
	}
	if err := c.rpc(ctx, "session/inject", map[string]any{"request": request}, &value); err != nil {
		return InjectResult{}, err
	}
	return InjectResult{Seqs: value.Seqs, Started: value.Started}, nil
}

// SetDraft stores text in the session composer without submitting a turn, so a
// caller can hand a prepared prompt to the operator for review. It is the
// transport behind Send(noSubmit) and the seeding step of a fork.
func (c Client) SetDraft(ctx context.Context, sessionID, text string) error {
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(text) == "" {
		return fmt.Errorf("DSH session and draft text are required")
	}
	request := map[string]any{"sessionId": sessionID, "text": text}
	return c.rpc(ctx, "session/setDraft", map[string]any{"request": request}, nil)
}

// Open opens the existing session, or does nothing for BrowserNone without
// discovering the server or invoking the OS opener. Service.Open is the explicit
// manual-opening operation. Opener errors redact the authenticated URL.
func (c Client) Open(sessionID, browser string) error {
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("DSH session ID is required")
	}
	browser, err := NormalizeBrowser(browser)
	if err != nil {
		return err
	}
	if browser == BrowserNone {
		return nil
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
