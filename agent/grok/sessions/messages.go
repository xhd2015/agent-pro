package sessions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// MessagesHelp is the text for `agent-pro grok session messages --help`.
const MessagesHelp = `Usage: agent-pro grok session messages (<session-id> | --tab SEL | --tab-index N) [OPTIONS]

Print the most recent coalesced Grok chat messages (msgfmt-style),
with per-kind rune caps (user 4096, tool 128, thinking 512, response 8192).
Each line is prefixed with a local [YYYY-MM-DD HH:MM:SS] timestamp, or [—]
when the wire time is unknown.

Session source (exactly one):
  <session-id>          explicit Grok session id
  --tab SEL             1-based tab index, or next|left|right (right ≡ next)
  --tab-index N         0-based tab index in this iTerm window

Options:
  --limit N             page size (default 32; 0 = all remaining after offset)
  --offset-from-end N   skip N newest messages before applying --limit (default 0)
                        example: --offset-from-end 32  # skip last 32; start next page
  --json                machine-readable (includes total, offset, limit; no ANSI)
  -h,--help             show help
`

// messagesMissingTimestampMarker matches grok prompts ([—], U+2014).
const messagesMissingTimestampMarker = "[—]"

const messagesTimestampLayout = "2006-01-02 15:04:05"

// MessagesCommandHelpLine is the parent `agent-pro grok session` help row.
const MessagesCommandHelpLine = `  messages …             print recent chat messages (--limit / --offset-from-end)`

// DefaultMessagesLimit is the default --limit when omitted.
const DefaultMessagesLimit = 32

// Per-kind body rune caps (msgfmt U+2026 ellipsis).
const (
	MessagesCapUser     = 4096
	MessagesCapTool     = 128
	MessagesCapThinking = 512
	MessagesCapResponse = 8192
)

// Chat message kinds for Messages / JSON.
const (
	MessageKindUser      = "user"
	MessageKindThinking  = "thinking"
	MessageKindTool      = "tool"
	MessageKindResponse  = "response"
)

// MessagesOpts drives Messages / RunMessages.
type MessagesOpts struct {
	Limit         int  // page size; 0 = all remaining after offset
	LimitSet      bool // true when --limit was explicitly passed
	OffsetFromEnd int
	JSON          bool

	// Loc formats timestamps. nil → time.Local.
	Loc *time.Location

	// Tab resolve hooks (nil → production). Used by RunMessages for --tab/--tab-index.
	ListProcs        func() []FocusProc
	Lsof             func(int) []string
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string
}

// ChatMessage is one coalesced transcript entry after kind caps.
type ChatMessage struct {
	Kind      string    `json:"kind"`
	Text      string    `json:"text"`
	Truncated bool      `json:"truncated"`
	Tool      string    `json:"tool,omitempty"`
	Timestamp time.Time `json:"-"` // wire time; zero = unknown
}

// MessagesResult is the outcome of Messages.
type MessagesResult struct {
	SessionID     string
	Total         int
	OffsetFromEnd int
	Limit         int // requested limit (0 = all remaining)
	Messages      []ChatMessage
	Text          string // formatted block (empty when no messages)
}

type messagesJSON struct {
	SessionID     string             `json:"session_id"`
	Total         int                `json:"total"`
	OffsetFromEnd int                `json:"offset_from_end"`
	Limit         int                `json:"limit"`
	Messages      []chatMessageJSON  `json:"messages"`
}

type chatMessageJSON struct {
	Kind      string `json:"kind"`
	Text      string `json:"text"`
	Truncated bool   `json:"truncated"`
	Tool      string `json:"tool,omitempty"`
	Timestamp string `json:"timestamp,omitempty"` // RFC3339 in Loc; omitted when unknown
}

// Messages loads coalesced chat messages for sessionID, applies offset/limit,
// and formats a msgfmt-style text block.
func Messages(grokHome, sessionID string, opts *MessagesOpts) (*MessagesResult, error) {
	if opts == nil {
		opts = &MessagesOpts{}
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	if opts.OffsetFromEnd < 0 {
		return nil, fmt.Errorf("--offset-from-end must be >= 0")
	}
	if opts.LimitSet && opts.Limit < 0 {
		return nil, fmt.Errorf("--limit must be >= 0")
	}

	info, err := Info(grokHome, sessionID)
	if err != nil {
		return nil, err
	}

	all, err := loadChatMessages(info.UpdatesPath)
	if err != nil {
		return nil, err
	}

	limit := opts.Limit
	if !opts.LimitSet {
		limit = DefaultMessagesLimit
	}

	page := pageMessagesFromEnd(all, opts.OffsetFromEnd, limit)
	loc := opts.Loc
	if loc == nil {
		loc = time.Local
	}
	text := ""
	if len(page) > 0 {
		text = formatChatMessagesText(page, len(all), loc)
	}

	return &MessagesResult{
		SessionID:     sessionID,
		Total:         len(all),
		OffsetFromEnd: opts.OffsetFromEnd,
		Limit:         limit,
		Messages:      page,
		Text:          text,
	}, nil
}

// RunMessages implements `agent-pro grok session messages` with injectable writers/hooks.
func RunMessages(args []string, stdout, stderr io.Writer, grokHome string, opts *MessagesOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parseMessagesArgs(args)
	if err != nil {
		return err
	}
	if parsed.Help {
		txt := MessagesHelp
		if !strings.HasSuffix(txt, "\n") {
			txt += "\n"
		}
		_, _ = io.WriteString(stdout, txt)
		return nil
	}

	runOpts := MessagesOpts{}
	if opts != nil {
		runOpts = *opts
	}
	runOpts.Limit = parsed.Limit
	runOpts.LimitSet = parsed.LimitSet
	runOpts.OffsetFromEnd = parsed.OffsetFromEnd
	runOpts.JSON = parsed.JSON

	sessionID, _, err := ResolveSessionSource(parsed.Positional, parsed.Tab, parsed.TabIndex, &SessionSourceOpts{
		ListProcs:        runOpts.ListProcs,
		Lsof:             runOpts.Lsof,
		ListITerm:        runOpts.ListITerm,
		CurrentSessionID: runOpts.CurrentSessionID,
		ControllingTTY:   runOpts.ControllingTTY,
		AncestorTTYs:     runOpts.AncestorTTYs,
	})
	if err != nil {
		return err
	}

	result, err := Messages(grokHome, sessionID, &runOpts)
	if err != nil {
		return err
	}

	if parsed.JSON {
		loc := runOpts.Loc
		if loc == nil {
			loc = time.Local
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(messagesJSON{
			SessionID:     result.SessionID,
			Total:         result.Total,
			OffsetFromEnd: result.OffsetFromEnd,
			Limit:         result.Limit,
			Messages:      chatMessagesToJSON(result.Messages, loc),
		}); err != nil {
			return err
		}
		_, err = stdout.Write(buf.Bytes())
		return err
	}

	if len(result.Messages) == 0 {
		_, err = io.WriteString(stdout, "(no messages)\n")
		return err
	}
	body := result.Text
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	_, err = io.WriteString(stdout, body)
	return err
}

type messagesArgs struct {
	Positional    []string
	Tab           *string
	TabIndex      *int
	Limit         int
	LimitSet      bool
	OffsetFromEnd int
	JSON          bool
	Help          bool
}

func parseMessagesArgs(args []string) (messagesArgs, error) {
	var out messagesArgs
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			out.Help = true
			return out, nil
		}
		if arg == "--json" {
			out.JSON = true
			continue
		}
		if arg == "--limit" || strings.HasPrefix(arg, "--limit=") {
			raw, next, err := takeFlagValue(arg, "--limit", args, &i)
			if err != nil {
				return out, err
			}
			if next {
				// i advanced by takeFlagValue
			}
			n, convErr := strconv.Atoi(raw)
			if convErr != nil {
				return out, fmt.Errorf("--limit must be an integer")
			}
			if n < 0 {
				return out, fmt.Errorf("--limit must be >= 0")
			}
			out.Limit = n
			out.LimitSet = true
			continue
		}
		if arg == "--offset-from-end" || strings.HasPrefix(arg, "--offset-from-end=") {
			raw, _, err := takeFlagValue(arg, "--offset-from-end", args, &i)
			if err != nil {
				return out, err
			}
			n, convErr := strconv.Atoi(raw)
			if convErr != nil {
				return out, fmt.Errorf("--offset-from-end must be an integer")
			}
			if n < 0 {
				return out, fmt.Errorf("--offset-from-end must be >= 0")
			}
			out.OffsetFromEnd = n
			continue
		}
		if arg == "--tab" || strings.HasPrefix(arg, "--tab=") {
			raw, _, err := takeFlagValue(arg, "--tab", args, &i)
			if err != nil {
				return out, err
			}
			out.Tab = &raw
			continue
		}
		if arg == "--tab-index" || strings.HasPrefix(arg, "--tab-index=") {
			raw, _, err := takeFlagValue(arg, "--tab-index", args, &i)
			if err != nil {
				return out, err
			}
			n, convErr := strconv.Atoi(raw)
			if convErr != nil {
				return out, fmt.Errorf("--tab-index must be an integer")
			}
			out.TabIndex = &n
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return out, fmt.Errorf("unknown flag: %s", arg)
		}
		out.Positional = append(out.Positional, arg)
	}
	return out, nil
}

func takeFlagValue(arg, name string, args []string, i *int) (string, bool, error) {
	if strings.HasPrefix(arg, name+"=") {
		return strings.TrimPrefix(arg, name+"="), false, nil
	}
	if *i+1 >= len(args) {
		return "", false, fmt.Errorf("%s requires a value", name)
	}
	*i++
	return args[*i], true, nil
}

func loadChatMessages(updatesPath string) ([]ChatMessage, error) {
	path := strings.TrimSpace(updatesPath)
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	events := grok_session.FromUpdatesJSONL(lines)
	return eventsToChatMessages(events), nil
}

func eventsToChatMessages(events []types.AgentEvent) []ChatMessage {
	out := make([]ChatMessage, 0, len(events))
	for _, ev := range events {
		msg, ok := eventToChatMessage(ev)
		if !ok {
			continue
		}
		out = append(out, msg)
	}
	return out
}

func eventToChatMessage(ev types.AgentEvent) (ChatMessage, bool) {
	ts := eventTimestamp(ev)
	switch ev.Type {
	case types.ActionMessage:
		role := strings.TrimSpace(ev.Role)
		text := ev.Text
		if text == "" {
			return ChatMessage{}, false
		}
		if role == "user" {
			body, trunc := capMessageBody(text, MessagesCapUser)
			return ChatMessage{Kind: MessageKindUser, Text: body, Truncated: trunc, Timestamp: ts}, true
		}
		body, trunc := capMessageBody(text, MessagesCapResponse)
		return ChatMessage{Kind: MessageKindResponse, Text: body, Truncated: trunc, Timestamp: ts}, true
	case types.ActionThink:
		if ev.Text == "" {
			return ChatMessage{}, false
		}
		body, trunc := capMessageBody(ev.Text, MessagesCapThinking)
		return ChatMessage{Kind: MessageKindThinking, Text: body, Truncated: trunc, Timestamp: ts}, true
	case types.ActionToolCall:
		// Prefer the invocation (pending). Completed/failed updates often only
		// add large Output and would duplicate the same call under a 128-rune cap.
		if st := grokToolStatus(ev); st == "completed" || st == "failed" {
			return ChatMessage{}, false
		}
		name := toolMessageName(ev)
		raw := formatToolUseBody(name, ev.ToolInput)
		if raw == "" {
			return ChatMessage{}, false
		}
		body, trunc := capMessageBody(raw, MessagesCapTool)
		return ChatMessage{Kind: MessageKindTool, Text: body, Truncated: trunc, Tool: name, Timestamp: ts}, true
	default:
		return ChatMessage{}, false
	}
}

func eventTimestamp(ev types.AgentEvent) time.Time {
	if ev.Timestamp <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ev.Timestamp)
}

func grokToolStatus(ev types.AgentEvent) string {
	if ev.Extensions == nil || ev.Extensions.GrokSession == nil {
		return ""
	}
	return strings.TrimSpace(ev.Extensions.GrokSession.Status)
}

func toolMessageName(ev types.AgentEvent) string {
	tool := strings.TrimSpace(ev.Tool)
	title := strings.TrimSpace(ev.Text)
	switch {
	case tool != "" && tool != "tool" && tool != "other":
		return tool
	case title != "":
		return title
	case tool != "":
		return tool
	default:
		return "tool"
	}
}

// formatToolUseBody builds a compact tool invocation line focused on the
// command / args (not tool output). Prefers ToolInput["command"].
func formatToolUseBody(name string, input map[string]any) string {
	name = strings.TrimSpace(name)
	detail := toolInputDetail(input)
	switch {
	case name != "" && detail != "":
		return name + ": " + detail
	case detail != "":
		return detail
	case name != "":
		return name
	default:
		return ""
	}
}

func toolInputDetail(input map[string]any) string {
	if len(input) == 0 {
		return ""
	}
	if cmd, ok := input["command"].(string); ok {
		if s := strings.TrimSpace(cmd); s != "" {
			return s
		}
	}
	// Prefer a few common single-arg shapes before dumping all keys.
	for _, key := range []string{"target_file", "path", "query", "pattern", "url"} {
		if v, ok := input[key].(string); ok {
			if s := strings.TrimSpace(v); s != "" {
				return key + "=" + s
			}
		}
	}
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+compactToolArg(input[k]))
	}
	return strings.Join(parts, " ")
}

func compactToolArg(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprint(x)
		}
		return string(b)
	}
}

// pageMessagesFromEnd skips offset newest messages, then keeps the last limit
// of what remains. limit <= 0 means all remaining.
func pageMessagesFromEnd(all []ChatMessage, offset, limit int) []ChatMessage {
	n := len(all)
	if n == 0 || offset >= n {
		return nil
	}
	end := n - offset
	start := 0
	if limit > 0 {
		start = end - limit
		if start < 0 {
			start = 0
		}
	}
	out := make([]ChatMessage, end-start)
	copy(out, all[start:end])
	return out
}

func formatChatMessagesText(page []ChatMessage, sourceCount int, loc *time.Location) string {
	if len(page) == 0 {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	var b strings.Builder
	b.WriteString(messagesHeader(len(page), sourceCount))
	b.WriteByte('\n')
	for _, m := range page {
		b.WriteString(formatMessageTimestamp(m.Timestamp, loc))
		b.WriteByte(' ')
		b.WriteString(fmt.Sprintf("[%s] : %s", kindSender(m.Kind), m.Text))
		b.WriteByte('\n')
	}
	return b.String()
}

func messagesHeader(shown, source int) string {
	if source == 1 && shown == 1 {
		return "Chat history (1 message):"
	}
	return fmt.Sprintf("Chat history (showing %d of %d):", shown, source)
}

func formatMessageTimestamp(ts time.Time, loc *time.Location) string {
	if ts.IsZero() {
		return messagesMissingTimestampMarker
	}
	if loc == nil {
		loc = time.Local
	}
	return "[" + ts.In(loc).Format(messagesTimestampLayout) + "]"
}

func chatMessagesToJSON(msgs []ChatMessage, loc *time.Location) []chatMessageJSON {
	if loc == nil {
		loc = time.Local
	}
	out := make([]chatMessageJSON, 0, len(msgs))
	for _, m := range msgs {
		row := chatMessageJSON{
			Kind:      m.Kind,
			Text:      m.Text,
			Truncated: m.Truncated,
			Tool:      m.Tool,
		}
		if !m.Timestamp.IsZero() {
			row.Timestamp = m.Timestamp.In(loc).Format(time.RFC3339)
		}
		out = append(out, row)
	}
	return out
}

func kindSender(kind string) string {
	switch kind {
	case MessageKindUser:
		return "user"
	case MessageKindThinking:
		return "thinking"
	case MessageKindTool:
		return "tool"
	case MessageKindResponse:
		return "assistant"
	default:
		return kind
	}
}

// capMessageBody shortens body so rune length is at most max.
// Truncated form is prefix of (max-1) runes + "…" (U+2026); result rune count == max.
func capMessageBody(body string, max int) (string, bool) {
	if max <= 0 {
		return body, false
	}
	n := utf8.RuneCountInString(body)
	if n <= max {
		return body, false
	}
	keep := max - 1
	if keep <= 0 {
		return "…", true
	}
	runes := []rune(body)
	return string(runes[:keep]) + "…", true
}
