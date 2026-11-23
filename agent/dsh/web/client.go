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
	// Kind selects the transcript role: "user", "thinking", "assistant",
	// "tool-call" or "tool-result".
	Kind string
	// Text is the message body. Blank text is rejected for the text kinds and
	// for tool-result; a tool-call carries Arguments instead.
	Text string
	// Tool is the tool name for a tool-call.
	Tool string
	// Arguments is the raw JSON argument object for a tool-call.
	Arguments string
	// Call is the call seq a tool-result answers.
	Call int64
	// IsError marks a failed tool-result.
	IsError bool
}

// InjectResult reports what an injected batch committed.
type InjectResult struct {
	// Seqs are the committed event sequence numbers, in batch order.
	Seqs []int64
	// CallSeqs are the committed tool/call sequence numbers, in batch order. A
	// call is a log event, so it appears here and not on the model-visible surface.
	CallSeqs []int64
	// ResultSeqs are the committed tool/result sequence numbers, in batch order.
	ResultSeqs []int64
	// Started reports whether the agent runs on the batch without another call.
	Started bool
}

// injectableKinds is the closed set DSH accepts, kept here so a typo fails
// before the round trip instead of surfacing as a server-side validation error.
var injectableKinds = map[string]bool{"user": true, "thinking": true, "assistant": true, "tool-call": true, "tool-result": true}

// RPCError is a refusal DSH itself produced, as opposed to a transport failure.
//
// Only the code is carried: the server's message can quote request content, and
// callers log what this renders.
type RPCError struct {
	// Method is the RPC that was refused, e.g. "session/edit".
	Method string
	// Code is DSH's own error code, e.g. "session/edit-rejected".
	Code string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("DSH RPC %s failed (%s)", e.Method, e.Code)
}

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
	wire := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		if !injectableKinds[message.Kind] {
			return InjectResult{}, fmt.Errorf("injected message kind must be user, thinking, assistant, tool-call or tool-result; got %q", message.Kind)
		}
		switch message.Kind {
		case "tool-call":
			if strings.TrimSpace(message.Tool) == "" {
				return InjectResult{}, fmt.Errorf("injected tool-call has no tool name")
			}
			if strings.TrimSpace(message.Arguments) == "" {
				return InjectResult{}, fmt.Errorf("injected tool-call for %q has no arguments", message.Tool)
			}
			wire = append(wire, map[string]any{"kind": message.Kind, "tool": message.Tool, "arguments": message.Arguments})
		case "tool-result":
			if strings.TrimSpace(message.Text) == "" {
				return InjectResult{}, fmt.Errorf("injected tool-result has no text")
			}
			entry := map[string]any{"kind": message.Kind, "call": message.Call, "text": message.Text}
			if message.IsError {
				entry["isError"] = true
			}
			wire = append(wire, entry)
		default:
			if strings.TrimSpace(message.Text) == "" {
				return InjectResult{}, fmt.Errorf("injected %s message has no text", message.Kind)
			}
			wire = append(wire, map[string]any{"kind": message.Kind, "text": message.Text})
		}
	}
	request := map[string]any{"sessionId": sessionID, "requestId": requestID, "messages": wire}
	var value struct {
		Seqs       []int64 `json:"seqs"`
		CallSeqs   []int64 `json:"callSeqs"`
		ResultSeqs []int64 `json:"resultSeqs"`
		Started    bool    `json:"started"`
	}
	if err := c.rpc(ctx, "session/inject", map[string]any{"request": request}, &value); err != nil {
		return InjectResult{}, err
	}
	return InjectResult{Seqs: value.Seqs, CallSeqs: value.CallSeqs, ResultSeqs: value.ResultSeqs, Started: value.Started}, nil
}

// EditedMessageChange is one change to an already-produced message.
type EditedMessageChange struct {
	// Seq names the message by its original surface seq; it is preserved.
	Seq int64
	// Text replaces the message content. Ignored when Delete is set.
	Text string
	// Delete removes the message from the model-visible history.
	Delete bool
}

// EditResult reports what one edit batch committed.
type EditResult struct {
	// Seq is the committed edit record.
	Seq int64
	// Started reports whether the agent runs on the edited history without
	// another call.
	Started bool
}

// EditMessages rewrites or removes messages the session already produced.
//
// Unlike injection this targets existing history, so every change must name a
// current message; the batch is one DSH transaction and an invalid change
// rejects the whole batch. A running agent picks the change up at its next step
// rather than mid-step, and an idle one is woken.
func (c Client) EditMessages(ctx context.Context, sessionID, requestID string, changes []EditedMessageChange) (EditResult, error) {
	if strings.TrimSpace(sessionID) == "" {
		return EditResult{}, fmt.Errorf("DSH session is required")
	}
	if strings.TrimSpace(requestID) == "" {
		return EditResult{}, fmt.Errorf("DSH request ID is required")
	}
	if len(changes) == 0 {
		return EditResult{}, fmt.Errorf("at least one edit is required")
	}
	wire := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		if change.Seq < 0 {
			return EditResult{}, fmt.Errorf("edit seq must be non-negative")
		}
		if change.Delete {
			wire = append(wire, map[string]any{"seq": change.Seq, "deleted": true})
			continue
		}
		if strings.TrimSpace(change.Text) == "" {
			return EditResult{}, fmt.Errorf("edit of seq %d has no text", change.Seq)
		}
		wire = append(wire, map[string]any{"seq": change.Seq, "text": change.Text})
	}
	request := map[string]any{"sessionId": sessionID, "requestId": requestID, "edits": wire}
	var value struct {
		Seq     int64 `json:"seq"`
		Started bool  `json:"started"`
	}
	if err := c.rpc(ctx, "session/edit", map[string]any{"request": request}, &value); err != nil {
		return EditResult{}, err
	}
	return EditResult{Seq: value.Seq, Started: value.Started}, nil
}

// SelectModel switches one Session's route without launching anything.
//
// It is the runtime counterpart of CreateOptions.ModelSelection: that one seeds a
// brand new Session, this one changes a Session that already exists, including
// one that is running right now. A running Session picks the new route up at its
// next step rather than mid-step, because a model swap inside one request would
// cut the reasoning and tool-call semantics it already committed to.
//
// The change is durable: DSH records it as a model/selection event, so closing
// and reopening the Session keeps it.
func (c Client) SelectModel(ctx context.Context, sessionID, provider, model, reasoningEffort string) (ModelSelection, error) {
	if strings.TrimSpace(sessionID) == "" {
		return ModelSelection{}, fmt.Errorf("DSH session is required")
	}
	next, err := normalizeModelSelection(&ModelSelection{
		Provider: provider, Model: model, ReasoningEffort: reasoningEffort,
	})
	if err != nil {
		return ModelSelection{}, err
	}
	request := map[string]any{
		"sessionId": sessionID,
		"provider":  next.Provider,
		"model":     next.Model,
	}
	if next.ReasoningEffort != "" {
		request["reasoningEffort"] = next.ReasoningEffort
	}
	var value struct {
		Selected *ModelSelection `json:"selected"`
	}
	if err := c.rpc(ctx, "session/selectModel", map[string]any{"request": request}, &value); err != nil {
		return ModelSelection{}, err
	}
	if value.Selected == nil {
		// An older Host may accept the switch without echoing it; the request we
		// sent is then the best available statement of what is now in force.
		return *next, nil
	}
	return *value.Selected, nil
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

// ArchiveOptions tunes one archive request.
type ArchiveOptions struct {
	// StopActivity stops the Session's running work -- its turn, subagent
	// descendants, owned background jobs and active schedules -- instead of
	// letting DSH refuse the archive as `workspace/session-active`.
	//
	// It defaults to false on purpose: archiving is a grouping decision, and
	// silently killing a working agent as a side effect of tidying a list is a
	// worse surprise than a refusal the caller can explain and retry.
	StopActivity bool
}

// ArchiveValue is DSH's complete archived-Session set after a mutation.
//
// It is the whole set, not the row just touched: DSH's archive is registry-wide,
// so a caller that keeps its own mirror can replace it instead of guessing.
type ArchiveValue struct {
	ArchivedSessionIDs []string
}

// archiveSessionRequest is the wire shape of workspace/archiveSession.
type archiveSessionRequest struct {
	SessionID    string `json:"sessionId"`
	StopActivity bool   `json:"stopActivity,omitempty"`
}

// Archive hides one Session from Workspace grouping surfaces.
//
// The Session stays readable; nothing on disk is deleted, and Unarchive brings it
// back. DSH refuses a Session with running work as `workspace/session-active`
// (returned here as *RPCError, whose Code names the refusal) unless
// ArchiveOptions.StopActivity is set, so a caller can decide whether stopping
// that work is part of its intent.
//
// One Session is one call. A thread that also has a verifier is two Sessions and
// therefore two calls: the verifier is not mounted on the main session.
func (c Client) Archive(ctx context.Context, sessionID string, opts ArchiveOptions) (ArchiveValue, error) {
	if strings.TrimSpace(sessionID) == "" {
		return ArchiveValue{}, fmt.Errorf("DSH session is required")
	}
	var value struct {
		ArchivedSessionIDs []string `json:"archivedSessionIds"`
	}
	request := archiveSessionRequest{SessionID: sessionID, StopActivity: opts.StopActivity}
	if err := c.rpc(ctx, "workspace/archiveSession", map[string]any{"request": request}, &value); err != nil {
		return ArchiveValue{}, err
	}
	return ArchiveValue{ArchivedSessionIDs: value.ArchivedSessionIDs}, nil
}

// Unarchive restores one archived Session to Workspace grouping surfaces.
//
// Idempotent: a Session that is not archived is not an error, so a caller that
// lost a race with another surface can simply retry.
func (c Client) Unarchive(ctx context.Context, sessionID string) (ArchiveValue, error) {
	if strings.TrimSpace(sessionID) == "" {
		return ArchiveValue{}, fmt.Errorf("DSH session is required")
	}
	var value struct {
		ArchivedSessionIDs []string `json:"archivedSessionIds"`
	}
	if err := c.rpc(ctx, "workspace/unarchiveSession", map[string]any{"request": map[string]any{"sessionId": sessionID}}, &value); err != nil {
		return ArchiveValue{}, err
	}
	return ArchiveValue{ArchivedSessionIDs: value.ArchivedSessionIDs}, nil
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
		// The failure is returned as a typed error so a caller can tell a refusal
		// it can explain (`edit-rejected`, `inject-failed`) from a transport
		// failure, without parsing this sentence. The server's own message stays
		// off the error string: it may echo request content, and callers log this.
		if envelope.Result.Error != nil {
			return &RPCError{Method: method, Code: envelope.Result.Error.Code}
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
