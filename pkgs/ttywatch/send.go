package ttywatch

import "strings"

// SendMessage prepares inject mode and sends verbatim bytes.
// When suffixCR is true, appends '\r' only if message contains no '\r'.
func SendMessage(listenAddr, sessionID, message string, suffixCR bool) error {
	if err := PrepareSessionInjectMode(listenAddr, sessionID); err != nil {
		return err
	}
	payload := message
	if suffixCR && !strings.Contains(payload, "\r") {
		payload += "\r"
	}
	return InjectInput(listenAddr, sessionID, []byte(payload))
}

// buildSendPayload returns the inject payload for SendMessage (test helper).
func buildSendPayload(message string, suffixCR bool) string {
	payload := message
	if suffixCR && !strings.Contains(payload, "\r") {
		payload += "\r"
	}
	return payload
}