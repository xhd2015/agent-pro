package agentsend

import "strings"

// normalizeSendText prepares user payload for follow-up delivery (send-queue /
// inject). Returns valid UTF-8 so runners never see truncated multi-byte
// sequences (same contract as argv open PROMPT for real grok std::env::args).
//
// Invalid sequences are replaced with U+FFFD so surrounding text is preserved.
func normalizeSendText(s string) string {
	return strings.ToValidUTF8(s, "\uFFFD")
}
