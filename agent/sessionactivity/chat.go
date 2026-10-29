package sessionactivity

// ClassifyChat maps a grok/codex last coalesced message onto the five
// activity kinds. live is PID liveness, not "the TUI is still open".
func ClassifyChat(lastKind, tool string, live bool) (Kind, string) {
	switch lastKind {
	case "thinking":
		if !live {
			return ErrorKind, "thinking but session is not live"
		}
		return Thinking, ""
	case "tool":
		if !live {
			return ErrorKind, "tool call but session is not live"
		}
		if tool == "ask_user_question" {
			return Waiting, tool
		}
		return ToolCalling, tool
	case "response":
		return Idle, ""
	case "user":
		if live {
			return Thinking, ""
		}
		return Idle, ""
	default:
		return Idle, ""
	}
}
