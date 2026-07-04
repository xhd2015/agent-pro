package ttywatch

import "regexp"

var (
	reCSI       = regexp.MustCompile(`\x1b\[[<]?[0-?]*[ -/]*[@-~]`)
	reOSC       = regexp.MustCompile(`\x1b\][^\x07]*(?:\x07|\x1b\\)`)
	reC0        = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
	reOrphanCSI = regexp.MustCompile(`\[[<]?(?:\d+;)+(?:\d+[A-Za-z~]?)?`)
)

// SanitizeForPrint strips CSI, OSC, and C0 control sequences except \n \r \t.
func SanitizeForPrint(data string) string {
	out := reCSI.ReplaceAllString(data, "")
	out = reOSC.ReplaceAllString(out, "")
	out = reC0.ReplaceAllString(out, "")
	out = reOrphanCSI.ReplaceAllString(out, "")
	return out
}