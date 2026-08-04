package agentsend

import "github.com/xhd2015/agent-pro/pkgs/utf8util"

// normalizeSendText prepares user payload for follow-up delivery (send-queue /
// inject). Returns valid UTF-8 so runners never see truncated multi-byte
// sequences (same contract as argv open PROMPT for real grok std::env::args).
func normalizeSendText(s string) string {
	return utf8util.ToValid(s)
}
