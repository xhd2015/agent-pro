package summary

import "strings"

func StringInputValue(input map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := input[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		if text = strings.TrimSpace(text); text != "" {
			return text
		}
	}
	return ""
}

func GenericToolInputSummary(input map[string]any) string {
	return StringInputValue(input, "doc", "url", "query", "pattern", "path", "command", "name")
}

func SkillInputSummary(input map[string]any) string {
	var parts []string
	if name := StringInputValue(input, "skill", "name"); name != "" {
		parts = append(parts, name)
	}
	if args := StringInputValue(input, "arguments", "args", "command"); args != "" {
		parts = append(parts, args)
	}
	return strings.Join(parts, "\n")
}

func TodoWriteInputSummary(input map[string]any) string {
	rawTodos, ok := input["todos"]
	if !ok {
		return ""
	}
	todos, ok := rawTodos.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, rawTodo := range todos {
		todo, ok := rawTodo.(map[string]any)
		if !ok {
			continue
		}
		content := StringInputValue(todo, "content", "text", "task")
		if content == "" {
			continue
		}
		if status := StringInputValue(todo, "status"); status != "" {
			parts = append(parts, status+": "+content)
		} else {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n")
}

func ToolInputSummary(tool string, input map[string]any) string {
	if len(input) == 0 {
		return ""
	}
	tool = strings.ToLower(strings.TrimSpace(tool))
	switch tool {
	case "bash", "shell", "execute", "exec", "run":
		return StringInputValue(input, "command", "cmd")
	case "read", "read_file", "read file":
		return StringInputValue(input, "filePath", "path", "file")
	case "write", "edit", "write_file", "write file", "patch":
		return StringInputValue(input, "filePath", "path", "file")
	case "glob", "grep", "search":
		return StringInputValue(input, "pattern", "query", "path")
	case "list", "ls", "list_files", "list files":
		return StringInputValue(input, "path", "pattern")
	case "webfetch":
		return StringInputValue(input, "url")
	case "websearch":
		return StringInputValue(input, "query", "search_term")
	case "skill":
		return SkillInputSummary(input)
	case "todowrite", "todo":
		return TodoWriteInputSummary(input)
	default:
		return GenericToolInputSummary(input)
	}
}