package hooks

import (
	"sort"

	"github.com/xhd2015/agent-pro/agent/opencode/config"
)

type Entry struct {
	Name        string
	Description string
	Template    string
}

type Info struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func Merge(data config.Data, entries []Entry) (added int, addedNames []string, existing []string) {
	cmd, ok := data["command"]
	if !ok {
		cmd = map[string]interface{}{}
		data["command"] = cmd
	}
	cmdMap, ok := cmd.(map[string]interface{})
	if !ok {
		cmdMap = map[string]interface{}{}
		data["command"] = cmdMap
	}

	for _, h := range entries {
		if _, exists := cmdMap[h.Name]; !exists {
			cmdMap[h.Name] = map[string]interface{}{
				"template":    h.Template,
				"description": h.Description,
			}
			added++
			addedNames = append(addedNames, h.Name)
		} else {
			existing = append(existing, h.Name)
		}
	}

	return
}

func Remove(cmdMap map[string]interface{}, entries []Entry) (removed int, removedNames []string) {
	for _, h := range entries {
		if _, exists := cmdMap[h.Name]; exists {
			delete(cmdMap, h.Name)
			removed++
			removedNames = append(removedNames, h.Name)
		}
	}
	return
}

func GetCommandHooks(data config.Data, names map[string]bool) []Info {
	cmd, ok := data["command"]
	if !ok {
		return nil
	}
	cmdMap, ok := cmd.(map[string]interface{})
	if !ok {
		return nil
	}

	var result []Info
	for name, val := range cmdMap {
		if !names[name] {
			continue
		}
		entry := Info{Name: name}
		cmdObj, ok := val.(map[string]interface{})
		if ok {
			if desc, ok := cmdObj["description"].(string); ok {
				entry.Description = desc
			}
		}
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
