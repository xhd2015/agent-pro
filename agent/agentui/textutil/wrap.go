package textutil

import "strings"

func WrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		for len(line) > width {
			cut := width
			for cut > 0 && line[cut] != ' ' {
				cut--
			}
			if cut == 0 {
				cut = width
			}
			lines = append(lines, line[:cut])
			if cut < len(line) && line[cut] == ' ' {
				line = line[cut+1:]
			} else {
				line = line[cut:]
			}
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func IndexByte(data []byte, b byte) int {
	for i, c := range data {
		if c == b {
			return i
		}
	}
	return -1
}
