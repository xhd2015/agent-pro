package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// MessagesHelp is the text for `agent-pro codex session messages --help`.
const MessagesHelp = `Usage: agent-pro codex session messages (<session-id> | --tab SEL | --tab-index N) [OPTIONS]

Print the most recent coalesced Codex chat messages (msgfmt-style),
with per-kind rune caps (user 4096, tool 128, thinking 512, response 8192).
Each line is prefixed with a local [YYYY-MM-DD HH:MM:SS] timestamp, or [—]
when the wire time is unknown. AGENTS/skills preambles are skipped.

Session source (exactly one):
  <session-id>          explicit Codex session id
  --tab SEL             1-based tab index, or next|left|right (right ≡ next)
  --tab-index N         0-based tab index in this iTerm window

Options:
  --limit N             page size (default 32; 0 = all remaining after offset)
  --offset-from-end N   skip N newest messages before applying --limit (default 0)
  --grep P              keep messages whose body contains P (repeatable; AND;
                        case-insensitive literal). Applied before offset/limit.
  --color               force ANSI color on (even when stdout is not a TTY)
  --no-color            force ANSI color off
  --json                machine-readable (includes total, offset, limit; no ANSI)
  -h,--help             show help
`

const MessagesCommandHelpLine = `  messages …             print recent chat messages (--limit / --grep / --offset-from-end)`

const (
	DefaultMessagesLimit = 32
	MessagesCapUser      = 4096
	MessagesCapTool      = 128
	MessagesCapThinking  = 512
	MessagesCapResponse  = 8192

	MessageKindUser     = "user"
	MessageKindThinking = "thinking"
	MessageKindTool     = "tool"
	MessageKindResponse = "response"

	messagesMissingTimestampMarker = "[—]"
	messagesTimestampLayout        = "2006-01-02 15:04:05"
)

// MessagesOpts drives Messages / RunMessages.
type MessagesOpts struct {
	Limit         int
	LimitSet      bool
	OffsetFromEnd int
	JSON          bool
	Greps         []string
	ColorMode     string // "auto" | "always" | "never"
	Loc           *time.Location

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
	Timestamp time.Time `json:"-"`
}

// MessagesResult is the outcome of Messages.
type MessagesResult struct {
	SessionID     string
	Total         int
	OffsetFromEnd int
	Limit         int
	Messages      []ChatMessage
	Text          string
}

type messagesJSON struct {
	SessionID     string            `json:"session_id"`
	Total         int               `json:"total"`
	OffsetFromEnd int               `json:"offset_from_end"`
	Limit         int               `json:"limit"`
	Messages      []chatMessageJSON `json:"messages"`
}

type chatMessageJSON struct {
	Kind      string `json:"kind"`
	Text      string `json:"text"`
	Truncated bool   `json:"truncated"`
	Tool      string `json:"tool,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// Messages loads coalesced chat messages for sessionID from the rollout JSONL.
func Messages(codexHome, sessionID string, opts *MessagesOpts) (*MessagesResult, error) {
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
	if err := validateMessagesGreps(opts.Greps); err != nil {
		return nil, err
	}

	path, err := Find(codexHome, sessionID)
	if err != nil {
		return nil, err
	}
	all, err := loadChatMessagesFromRollout(path)
	if err != nil {
		return nil, err
	}

	limit := opts.Limit
	if !opts.LimitSet {
		limit = DefaultMessagesLimit
	}
	filtered := all
	if len(opts.Greps) > 0 {
		filtered = filterMessagesByGrep(all, opts.Greps)
	}
	page := pageMessagesFromEnd(filtered, opts.OffsetFromEnd, limit)
	loc := opts.Loc
	if loc == nil {
		loc = time.Local
	}
	text := ""
	if len(page) > 0 {
		text = formatChatMessagesText(page, len(filtered), opts.OffsetFromEnd, loc)
	}
	return &MessagesResult{
		SessionID:     sessionID,
		Total:         len(filtered),
		OffsetFromEnd: opts.OffsetFromEnd,
		Limit:         limit,
		Messages:      page,
		Text:          text,
	}, nil
}

// RunMessages implements `agent-pro codex session messages`.
func RunMessages(args []string, stdout, stderr io.Writer, codexHome string, opts *MessagesOpts) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	parsed, err := parseMessagesArgs(args)
	if err != nil {
		return fmt.Errorf("agent-pro codex session messages: %w", err)
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
	runOpts.Greps = parsed.Greps
	runOpts.ColorMode = parsed.ColorMode

	sessionID, _, err := ResolveSessionSource(parsed.Positional, parsed.Tab, parsed.TabIndex, &SessionSourceOpts{
		ListProcs:        runOpts.ListProcs,
		Lsof:             runOpts.Lsof,
		ListITerm:        runOpts.ListITerm,
		CurrentSessionID: runOpts.CurrentSessionID,
		ControllingTTY:   runOpts.ControllingTTY,
		AncestorTTYs:     runOpts.AncestorTTYs,
	})
	if err != nil {
		return fmt.Errorf("agent-pro codex session messages: %w", err)
	}

	result, err := Messages(codexHome, sessionID, &runOpts)
	if err != nil {
		return err
	}

	loc := runOpts.Loc
	if loc == nil {
		loc = time.Local
	}

	if parsed.JSON {
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
		empty := "(no messages)\n"
		if len(runOpts.Greps) > 0 {
			empty = "(no matching messages)\n"
		}
		_, err = io.WriteString(stdout, empty)
		return err
	}
	_ = runOpts.ColorMode // color reserved; human output stays plain for stability
	return writeChatMessages(stdout, result.Messages, result.Total, result.OffsetFromEnd, loc, false, runOpts.Greps)
}

type messagesArgs struct {
	Positional    []string
	Tab           *string
	TabIndex      *int
	Limit         int
	LimitSet      bool
	OffsetFromEnd int
	JSON          bool
	Greps         []string
	ColorMode     string
	Help          bool
}

func parseMessagesArgs(args []string) (messagesArgs, error) {
	var out messagesArgs
	var colorFlag, noColorFlag bool
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
		if arg == "--color" {
			colorFlag = true
			continue
		}
		if arg == "--no-color" {
			noColorFlag = true
			continue
		}
		if arg == "--grep" || strings.HasPrefix(arg, "--grep=") {
			raw, _, err := takeMessagesFlagValue(arg, "--grep", args, &i)
			if err != nil {
				return out, err
			}
			out.Greps = append(out.Greps, raw)
			continue
		}
		if arg == "--limit" || strings.HasPrefix(arg, "--limit=") {
			raw, _, err := takeMessagesFlagValue(arg, "--limit", args, &i)
			if err != nil {
				return out, err
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
			raw, _, err := takeMessagesFlagValue(arg, "--offset-from-end", args, &i)
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
			raw, _, err := takeMessagesFlagValue(arg, "--tab", args, &i)
			if err != nil {
				return out, err
			}
			out.Tab = &raw
			continue
		}
		if arg == "--tab-index" || strings.HasPrefix(arg, "--tab-index=") {
			raw, _, err := takeMessagesFlagValue(arg, "--tab-index", args, &i)
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
	if colorFlag && noColorFlag {
		return out, fmt.Errorf("--color and --no-color cannot be specified together")
	}
	switch {
	case colorFlag:
		out.ColorMode = "always"
	case noColorFlag:
		out.ColorMode = "never"
	default:
		out.ColorMode = "auto"
	}
	if err := validateMessagesGreps(out.Greps); err != nil {
		return out, err
	}
	return out, nil
}

func takeMessagesFlagValue(arg, name string, args []string, i *int) (string, bool, error) {
	if strings.HasPrefix(arg, name+"=") {
		return strings.TrimPrefix(arg, name+"="), false, nil
	}
	if *i+1 >= len(args) {
		return "", false, fmt.Errorf("%s requires a value", name)
	}
	*i++
	return args[*i], true, nil
}

func validateMessagesGreps(greps []string) error {
	for _, g := range greps {
		if g == "" {
			return fmt.Errorf("--grep pattern must not be empty")
		}
	}
	return nil
}

func filterMessagesByGrep(msgs []ChatMessage, greps []string) []ChatMessage {
	if len(greps) == 0 {
		return msgs
	}
	out := make([]ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		ok := true
		lower := strings.ToLower(m.Text)
		for _, g := range greps {
			if !strings.Contains(lower, strings.ToLower(g)) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, m)
		}
	}
	return out
}

func loadChatMessagesFromRollout(path string) ([]ChatMessage, error) {
	path = strings.TrimSpace(path)
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

	sc := newRolloutScanner(f)
	var out []ChatMessage
	for sc.Scan() {
		line := sc.Text()
		msg, ok := chatMessageFromRolloutLine(line)
		if !ok {
			continue
		}
		out = append(out, msg)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func chatMessageFromRolloutLine(line string) (ChatMessage, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ChatMessage{}, false
	}
	var row struct {
		Type      string          `json:"type"`
		Timestamp string          `json:"timestamp"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal([]byte(trimmed), &row); err != nil {
		return ChatMessage{}, false
	}
	ts := parseRolloutTimestamp(row.Timestamp)

	switch row.Type {
	case "event_msg":
		var p eventMsgPayload
		if err := json.Unmarshal(row.Payload, &p); err != nil {
			return ChatMessage{}, false
		}
		switch p.Type {
		case "agent_message":
			if p.Phase != "commentary" && p.Phase != "final" && p.Phase != "" {
				return ChatMessage{}, false
			}
			text := strings.TrimSpace(p.Message)
			if text == "" {
				return ChatMessage{}, false
			}
			body, trunc := capMessageBody(text, MessagesCapResponse)
			return ChatMessage{Kind: MessageKindResponse, Text: body, Truncated: trunc, Timestamp: ts}, true
		case "user_message":
			text := strings.TrimSpace(p.Message)
			if text == "" || isTitlePreamble(text) {
				return ChatMessage{}, false
			}
			body, trunc := capMessageBody(text, MessagesCapUser)
			return ChatMessage{Kind: MessageKindUser, Text: body, Truncated: trunc, Timestamp: ts}, true
		default:
			return ChatMessage{}, false
		}
	case "response_item":
		var p struct {
			Type      string          `json:"type"`
			Role      string          `json:"role"`
			Name      string          `json:"name"`
			Arguments string          `json:"arguments"`
			Input     string          `json:"input"`
			Text      string          `json:"text"`
			Summary   string          `json:"summary"`
			Content   json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(row.Payload, &p); err != nil {
			return ChatMessage{}, false
		}
		switch p.Type {
		case "message":
			role := strings.ToLower(strings.TrimSpace(p.Role))
			if role == "developer" || role == "system" {
				return ChatMessage{}, false
			}
			text := strings.TrimSpace(p.Text)
			if text == "" {
				text = strings.TrimSpace(textFromMessageContent(p.Content))
			}
			if text == "" {
				return ChatMessage{}, false
			}
			if role == "user" {
				if isTitlePreamble(text) {
					return ChatMessage{}, false
				}
				body, trunc := capMessageBody(text, MessagesCapUser)
				return ChatMessage{Kind: MessageKindUser, Text: body, Truncated: trunc, Timestamp: ts}, true
			}
			// assistant (and other non-system roles)
			body, trunc := capMessageBody(text, MessagesCapResponse)
			return ChatMessage{Kind: MessageKindResponse, Text: body, Truncated: trunc, Timestamp: ts}, true
		case "reasoning":
			text := strings.TrimSpace(p.Text)
			if text == "" {
				text = strings.TrimSpace(p.Summary)
			}
			if text == "" {
				return ChatMessage{}, false
			}
			body, trunc := capMessageBody(text, MessagesCapThinking)
			return ChatMessage{Kind: MessageKindThinking, Text: body, Truncated: trunc, Timestamp: ts}, true
		case "function_call":
			name := strings.TrimSpace(p.Name)
			if name == "" {
				name = "tool"
			}
			detail := strings.TrimSpace(p.Arguments)
			if detail == "" {
				detail = strings.TrimSpace(p.Input)
			}
			raw := name
			if detail != "" {
				raw = name + " " + detail
			}
			body, trunc := capMessageBody(raw, MessagesCapTool)
			return ChatMessage{Kind: MessageKindTool, Text: body, Truncated: trunc, Tool: name, Timestamp: ts}, true
		case "custom_tool_call":
			name := strings.TrimSpace(p.Name)
			if name == "" {
				name = "tool"
			}
			detail := strings.TrimSpace(p.Input)
			raw := name
			if detail != "" {
				raw = name + " " + detail
			}
			body, trunc := capMessageBody(raw, MessagesCapTool)
			return ChatMessage{Kind: MessageKindTool, Text: body, Truncated: trunc, Tool: name, Timestamp: ts}, true
		default:
			return ChatMessage{}, false
		}
	default:
		return ChatMessage{}, false
	}
}

func parseRolloutTimestamp(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t
	}
	return time.Time{}
}

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

func formatChatMessagesText(page []ChatMessage, sourceCount, offsetFromEnd int, loc *time.Location) string {
	if len(page) == 0 {
		return ""
	}
	var b strings.Builder
	_ = writeChatMessages(&b, page, sourceCount, offsetFromEnd, loc, false, nil)
	return b.String()
}

func writeChatMessages(w io.Writer, page []ChatMessage, sourceCount, offsetFromEnd int, loc *time.Location, useColor bool, greps []string) error {
	_ = useColor
	_ = greps
	if len(page) == 0 {
		return nil
	}
	if loc == nil {
		loc = time.Local
	}
	if _, err := fmt.Fprintln(w, messagesHeader(len(page), sourceCount, offsetFromEnd)); err != nil {
		return err
	}
	for _, m := range page {
		if _, err := fmt.Fprintln(w, formatChatMessageLine(m, loc)); err != nil {
			return err
		}
	}
	return nil
}

func formatChatMessageLine(m ChatMessage, loc *time.Location) string {
	ts := formatMessageTimestamp(m.Timestamp, loc)
	return ts + " " + fmt.Sprintf("[%s] : %s", kindSender(m.Kind), m.Text)
}

func messagesHeader(shown, source, offsetFromEnd int) string {
	if source == 1 && shown == 1 {
		return "Chat history (1 message):"
	}
	if offsetFromEnd <= 0 {
		if shown == source {
			return fmt.Sprintf("Chat history (showing all %d of %d):", shown, source)
		}
		return fmt.Sprintf("Chat history (showing last %d of %d):", shown, source)
	}
	hi := source - offsetFromEnd
	lo := hi - shown + 1
	return fmt.Sprintf("Chat history (showing %d-%d(%d) of %d):", lo, hi, shown, source)
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
