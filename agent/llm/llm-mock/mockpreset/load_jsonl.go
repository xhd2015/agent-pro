package mockpreset

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// LoadJSONL reads AgentEvent JSON lines into a genQueue prefix.
// Empty lines are skipped. Missing path is an error.
func LoadJSONL(path string) ([]types.AgentEvent, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("empty mock-events-file path")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []types.AgentEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var evt types.AgentEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}
		if evt.Type == "" {
			return nil, fmt.Errorf("%s:%d: missing type", path, lineNo)
		}
		out = append(out, evt)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
