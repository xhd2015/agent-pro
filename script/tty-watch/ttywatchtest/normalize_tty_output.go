package ttywatchtest

import "bytes"

// NormalizeTTYOutput mirrors pkgs/ttywatch/attach_client.go normalizeTTYOutput for doctest
// regression locks on raw TTY CRLF shaping (bare LF → CRLF; standalone CR preserved).
func NormalizeTTYOutput(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	var out bytes.Buffer
	out.Grow(len(b) + bytes.Count(b, []byte{'\n'}))
	for i, c := range b {
		if c != '\n' {
			out.WriteByte(c)
			continue
		}
		if i > 0 && b[i-1] == '\r' {
			out.WriteByte('\n')
			continue
		}
		out.WriteString("\r\n")
	}
	return out.Bytes()
}