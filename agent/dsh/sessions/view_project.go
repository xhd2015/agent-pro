package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Project turns committed events into preview rows (thinking, present, ask,
// tool summaries, turn-end errors). It does not filter afterSeq/tail.
func Project(events []Event) []Message {
	out := make([]Message, 0, len(events))
	for _, event := range events {
		out = append(out, projectEvent(event)...)
	}
	byCall := map[string]*Ask{}
	for i := range out {
		if out[i].Ask != nil && out[i].Ask.CallID != "" {
			byCall[out[i].Ask.CallID] = out[i].Ask
		}
	}
	for _, event := range events {
		callID, answers := askResult(event)
		if callID == "" || len(answers) == 0 {
			continue
		}
		if row := byCall[callID]; row != nil {
			row.Answers = answers
		}
	}
	return out
}

func projectEvent(event Event) []Message {
	if event.Type == "turn/end" {
		return turnEndRows(event)
	}
	if event.Type == "assistant/message" {
		return assistantRows(event)
	}
	rows := Transcript([]Event{event}, 0, 0)
	if len(rows) == 0 {
		return nil
	}
	out := make([]Message, 0, len(rows))
	for _, row := range rows {
		if event.Type == "tool/result" && row.Tool == "ask_user_question" {
			continue
		}
		message := Message{Seq: row.Seq, Time: row.Time, Kind: row.Kind, Text: row.Text, Tool: row.Tool}
		message.Present = messagePresent(event, row)
		message.Ask = messageAsk(event, row)
		message.Sum = toolSummary(event, row)
		if message.Ask != nil && message.Sum == "" && len(message.Ask.Questions) > 0 {
			message.Sum = sumLine(message.Ask.Questions[0].Question)
		}
		if row.Kind == TranscriptUser {
			message.Images = userImageRefs(event)
		}
		out = append(out, message)
	}
	return out
}

func messagePresent(event Event, row TranscriptRow) *Present {
	if row.Kind != TranscriptTool || row.Tool != "present" {
		return nil
	}
	var envelope struct {
		Arguments json.RawMessage `json:"arguments"`
	}
	_ = json.Unmarshal(event.Data, &envelope)
	files := presentFiles(envelope.Arguments)
	if len(files) == 0 {
		files = presentFiles(event.Data)
	}
	if len(files) == 0 {
		return nil
	}
	return &Present{Files: files}
}

func messageAsk(event Event, row TranscriptRow) *Ask {
	if row.Kind != TranscriptTool || row.Tool != "ask_user_question" {
		return nil
	}
	var envelope struct {
		CallID    string          `json:"callId"`
		Arguments json.RawMessage `json:"arguments"`
	}
	_ = json.Unmarshal(event.Data, &envelope)
	questions := askQuestions(envelope.Arguments)
	if len(questions) == 0 {
		questions = askQuestions(event.Data)
	}
	if len(questions) == 0 {
		return nil
	}
	return &Ask{CallID: strings.TrimSpace(envelope.CallID), Questions: questions}
}

func askQuestions(raw json.RawMessage) []AskQuestion {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	if raw[0] == '"' {
		var inner string
		if json.Unmarshal(raw, &inner) != nil {
			return nil
		}
		return askQuestions([]byte(inner))
	}
	var obj struct {
		Questions []struct {
			ID          string      `json:"id"`
			Question    string      `json:"question"`
			Header      string      `json:"header,omitempty"`
			Detail      string      `json:"detail,omitempty"`
			Options     []AskOption `json:"options,omitempty"`
			MultiSelect bool        `json:"multiSelect"`
			MultiSnake  bool        `json:"multi_select"`
		} `json:"questions"`
	}
	if json.Unmarshal(raw, &obj) != nil || len(obj.Questions) == 0 {
		return nil
	}
	out := make([]AskQuestion, 0, len(obj.Questions))
	for _, item := range obj.Questions {
		question := strings.TrimSpace(item.Question)
		if question == "" {
			continue
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = fmt.Sprintf("q%d", len(out)+1)
		}
		out = append(out, AskQuestion{
			ID: id, Question: question, Header: strings.TrimSpace(item.Header),
			Detail: strings.TrimSpace(item.Detail), Options: item.Options,
			MultiSelect: item.MultiSelect || item.MultiSnake,
		})
	}
	return out
}

func turnEndRows(event Event) []Message {
	kind, message, code := turnEndReason(event.Data)
	if kind != "error" && kind != "aborted" {
		return nil
	}
	text := strings.TrimSpace(message)
	if text == "" {
		text = kind
	}
	outKind := "error"
	if kind == "aborted" {
		outKind = "aborted"
	}
	return []Message{{Seq: event.Seq, Time: event.Time, Kind: outKind, Text: text, Sum: code}}
}

func turnEndReason(raw json.RawMessage) (kind, message, code string) {
	var data map[string]json.RawMessage
	if json.Unmarshal(raw, &data) != nil || data == nil {
		return "", "", ""
	}
	var reason map[string]json.RawMessage
	if json.Unmarshal(data["reason"], &reason) != nil || reason == nil {
		return "", "", ""
	}
	_ = json.Unmarshal(reason["kind"], &kind)
	kind = strings.TrimSpace(kind)
	if msg := strings.TrimSpace(jsonString(reason["message"])); msg != "" {
		message = msg
	}
	var nested map[string]json.RawMessage
	if json.Unmarshal(reason["error"], &nested) == nil && nested != nil {
		if msg := strings.TrimSpace(jsonString(nested["message"])); msg != "" {
			message = msg
		}
		if c := strings.TrimSpace(jsonString(nested["code"])); c != "" {
			code = c
		}
	} else if msg := strings.TrimSpace(jsonString(reason["error"])); msg != "" && message == "" {
		message = msg
	}
	if c := strings.TrimSpace(jsonString(reason["code"])); c != "" && code == "" {
		code = c
	}
	return kind, message, code
}

func jsonString(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	return ""
}

func askResult(event Event) (string, []AskAnswer) {
	if event.Type != "tool/result" {
		return "", nil
	}
	var envelope struct {
		CallID  string          `json:"callId"`
		Name    string          `json:"name"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal(event.Data, &envelope) != nil {
		return "", nil
	}
	if strings.TrimSpace(envelope.Name) != "" && envelope.Name != "ask_user_question" {
		return "", nil
	}
	text := toolResultText(envelope.Content)
	if text == "" {
		return strings.TrimSpace(envelope.CallID), nil
	}
	var payload struct {
		Answers []AskAnswer `json:"answers"`
	}
	if json.Unmarshal([]byte(text), &payload) != nil || len(payload.Answers) == 0 {
		return strings.TrimSpace(envelope.CallID), nil
	}
	return strings.TrimSpace(envelope.CallID), payload.Answers
}

func toolResultText(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return ""
	}
	if raw[0] == '"' {
		var text string
		if json.Unmarshal(raw, &text) == nil {
			return strings.TrimSpace(text)
		}
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) == nil {
		for _, block := range blocks {
			if block.Type == "text" || block.Type == "" {
				if text := strings.TrimSpace(block.Text); text != "" {
					return text
				}
			}
		}
	}
	return ""
}

func presentFiles(raw json.RawMessage) []PresentFile {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	if raw[0] == '"' {
		var inner string
		if json.Unmarshal(raw, &inner) != nil {
			return nil
		}
		return presentFiles([]byte(inner))
	}
	var obj struct {
		Files []PresentFile `json:"files"`
		Path  string        `json:"path"`
		File  string        `json:"file"`
	}
	if json.Unmarshal(raw, &obj) != nil {
		return nil
	}
	if len(obj.Files) > 0 {
		return obj.Files
	}
	if path := strings.TrimSpace(obj.Path); path != "" {
		return []PresentFile{{Path: path}}
	}
	if path := strings.TrimSpace(obj.File); path != "" {
		return []PresentFile{{Path: path}}
	}
	return nil
}

func toolArguments(event Event) map[string]json.RawMessage {
	var data struct {
		Arguments json.RawMessage `json:"arguments"`
	}
	if json.Unmarshal(event.Data, &data) != nil || len(bytes.TrimSpace(data.Arguments)) == 0 {
		return nil
	}
	raw := bytes.TrimSpace(data.Arguments)
	if len(raw) > 0 && raw[0] == '"' {
		var encoded string
		if json.Unmarshal(raw, &encoded) != nil {
			return nil
		}
		raw = []byte(encoded)
	}
	var args map[string]json.RawMessage
	if json.Unmarshal(raw, &args) != nil {
		return nil
	}
	return args
}

func toolSummary(event Event, row TranscriptRow) string {
	args := toolArguments(event)
	arg := func(key string) string {
		var value string
		if json.Unmarshal(args[key], &value) != nil {
			return ""
		}
		return strings.TrimSpace(value)
	}
	anyArg := func(keys ...string) string {
		for _, key := range keys {
			if value := arg(key); value != "" {
				return value
			}
		}
		return ""
	}
	switch row.Tool {
	case "read", "write", "edit":
		if path := anyArg("path", "file_path", "notebook_path"); path != "" {
			return sumLine(path)
		}
	}
	if row.Text != "" {
		return sumLine(firstLine(row.Text))
	}
	if args == nil {
		return ""
	}
	switch row.Tool {
	case "glob":
		if pattern := arg("pattern"); pattern != "" {
			if path := arg("path"); path != "" {
				return pattern + " in " + path
			}
			return pattern
		}
	case "grep":
		if pattern := arg("pattern"); pattern != "" {
			return pattern
		}
	case "todo_write":
		var todos []struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(args["todos"], &todos); err == nil && len(todos) > 0 {
			done := 0
			for _, todo := range todos {
				if todo.Status == "completed" {
					done++
				}
			}
			return fmt.Sprintf("%d todos · %d done", len(todos), done)
		}
	case "job_output":
		if id := arg("jobId"); id != "" {
			return "job " + id
		}
	case "web_fetch":
		return anyArg("url")
	case "web_search":
		return anyArg("query")
	case "task", "agent":
		return anyArg("description", "prompt")
	}
	for _, raw := range args {
		var value string
		if json.Unmarshal(raw, &value) == nil && strings.TrimSpace(value) != "" {
			return sumLine(firstLine(value))
		}
	}
	return ""
}

func sumLine(text string) string {
	const maxSum = 160
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > maxSum {
		return string(runes[:maxSum]) + "…"
	}
	return string(runes)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func eventContent(event Event) []map[string]json.RawMessage {
	var data map[string]json.RawMessage
	if json.Unmarshal(event.Data, &data) != nil || data == nil {
		return nil
	}
	raw := data["content"]
	if event.Type == "assistant/message" {
		if msg, ok := data["message"]; ok {
			var inner map[string]json.RawMessage
			if json.Unmarshal(msg, &inner) == nil && inner["content"] != nil {
				raw = inner["content"]
			}
		}
	}
	var blocks []map[string]json.RawMessage
	if json.Unmarshal(raw, &blocks) != nil {
		return nil
	}
	return blocks
}

func blockString(block map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		var value string
		if json.Unmarshal(block[key], &value) == nil && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func userImageRefs(event Event) []Image {
	if event.Type != "user/message" {
		return nil
	}
	var out []Image
	index := 0
	for _, block := range eventContent(event) {
		mime, ok := userImageMeta(block)
		if !ok {
			continue
		}
		out = append(out, Image{MimeType: mime, Index: index})
		index++
	}
	return out
}

func userImageMeta(block map[string]json.RawMessage) (string, bool) {
	if blockString(block, "type") != "image" {
		return "", false
	}
	id, mime := blockAttachment(block)
	if blockString(block, "data") == "" && id == "" {
		return "", false
	}
	if mime == "" {
		mime = blockString(block, "mediaType", "mimeType")
	}
	if mime == "" {
		mime = "image/png"
	}
	return mime, true
}

func blockAttachment(block map[string]json.RawMessage) (id, mime string) {
	var att struct {
		AttachmentID string `json:"attachmentId"`
		MediaType    string `json:"mediaType"`
	}
	if json.Unmarshal(block["attachment"], &att) != nil {
		return "", ""
	}
	id = strings.TrimSpace(att.AttachmentID)
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) != 64 {
		return "", strings.TrimSpace(att.MediaType)
	}
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", strings.TrimSpace(att.MediaType)
		}
	}
	return id, strings.TrimSpace(att.MediaType)
}

type contentBlock struct {
	Type string
	Text string
}

func assistantRows(event Event) []Message {
	var out []Message
	var text strings.Builder
	flushText := func() {
		if strings.TrimSpace(text.String()) == "" {
			text.Reset()
			return
		}
		out = append(out, Message{Seq: event.Seq, Time: event.Time, Kind: "assistant", Text: text.String()})
		text.Reset()
	}
	for _, block := range assistantBlocks(event) {
		switch block.Type {
		case "reasoning", "thinking":
			flushText()
			if strings.TrimSpace(block.Text) == "" {
				continue
			}
			out = append(out, Message{Seq: event.Seq, Time: event.Time, Kind: "thinking", Text: block.Text, Sum: sumLine(block.Text)})
		case "text":
			text.WriteString(block.Text)
		}
	}
	flushText()
	return out
}

func assistantBlocks(event Event) []contentBlock {
	var data map[string]json.RawMessage
	if json.Unmarshal(event.Data, &data) != nil || data == nil {
		return nil
	}
	raw := data["content"]
	if msg, ok := data["message"]; ok {
		var inner map[string]json.RawMessage
		if json.Unmarshal(msg, &inner) == nil && inner != nil {
			if _, ok := inner["content"]; ok {
				raw = inner["content"]
			}
		}
	}
	var blocks []map[string]json.RawMessage
	if json.Unmarshal(raw, &blocks) != nil {
		return nil
	}
	out := make([]contentBlock, 0, len(blocks))
	for _, block := range blocks {
		var kind string
		if json.Unmarshal(block["type"], &kind) != nil {
			continue
		}
		value := blockText(block)
		if value == "" && kind != "text" {
			continue
		}
		out = append(out, contentBlock{Type: kind, Text: value})
	}
	return out
}

func blockText(block map[string]json.RawMessage) string {
	for _, key := range []string{"text", "reasoning"} {
		var value *string
		if json.Unmarshal(block[key], &value) != nil || value == nil {
			continue
		}
		if strings.TrimSpace(*value) != "" {
			return *value
		}
	}
	return ""
}

// eventAgentTodos returns the list written by a todo_write tool call. Callers
// that fold events one at a time keep the last ok result instead of re-scanning
// every event.
func eventAgentTodos(event Event) ([]Todo, bool) {
	if event.Type != "tool/call" {
		return nil, false
	}
	var data struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(event.Data, &data) != nil || data.Name != "todo_write" {
		return nil, false
	}
	args := toolArguments(event)
	var todos []Todo
	if json.Unmarshal(args["todos"], &todos) != nil {
		return nil, false
	}
	kept := make([]Todo, 0, len(todos))
	for _, todo := range todos {
		content := strings.TrimSpace(todo.Content)
		if content == "" {
			continue
		}
		status := strings.TrimSpace(todo.Status)
		if status == "" {
			status = "pending"
		}
		kept = append(kept, Todo{Content: content, Status: status})
	}
	return kept, true
}

func latestAgentTodos(events []Event) []Todo {
	var out []Todo
	for _, event := range events {
		if todos, ok := eventAgentTodos(event); ok {
			out = todos
		}
	}
	return out
}

// eventReasoningEffort returns the effort a request/header event names, or ""
// when the event carries none.
func eventReasoningEffort(event Event) string {
	if event.Type != "request/header" {
		return ""
	}
	var data struct {
		Header struct {
			Config struct {
				ReasoningEffort string `json:"reasoningEffort"`
			} `json:"config"`
		} `json:"header"`
	}
	if json.Unmarshal(event.Data, &data) != nil {
		return ""
	}
	return strings.TrimSpace(data.Header.Config.ReasoningEffort)
}

func lastReasoningEffort(events []Event) string {
	effort := ""
	for _, event := range events {
		if next := eventReasoningEffort(event); next != "" {
			effort = next
		}
	}
	return effort
}
