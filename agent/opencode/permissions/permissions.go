package permissions

import "github.com/xhd2015/agent-traces/agent/opencode/config"

type DenyRule struct {
	Pattern string
	Reason  string
}

func GetBash(data config.Data) interface{} {
	perm, ok := data["permission"]
	if !ok {
		return nil
	}
	permMap, ok := perm.(map[string]interface{})
	if !ok {
		return nil
	}
	bash, ok := permMap["bash"]
	if !ok {
		return nil
	}
	return bash
}

func MergeDenyRules(data config.Data, rules []DenyRule) (added int, addedPatterns []string, existing []string) {
	perm, ok := data["permission"]
	if !ok {
		perm = map[string]interface{}{}
		data["permission"] = perm
	}
	permMap, ok := perm.(map[string]interface{})
	if !ok {
		permMap = map[string]interface{}{}
		data["permission"] = permMap
	}

	bash, ok := permMap["bash"]
	if !ok {
		bash = map[string]interface{}{}
		permMap["bash"] = bash
	}

	bashMap, ok := bash.(map[string]interface{})
	if !ok {
		bashMap = map[string]interface{}{}
		permMap["bash"] = bashMap
	}

	for _, rule := range rules {
		if _, exists := bashMap[rule.Pattern]; !exists {
			bashMap[rule.Pattern] = "deny"
			added++
			addedPatterns = append(addedPatterns, rule.Pattern)
		} else {
			existing = append(existing, rule.Pattern)
		}
	}

	return
}

func RemoveDenyRules(bashMap map[string]interface{}, rules []DenyRule) (removed int, removedPatterns []string) {
	for _, rule := range rules {
		if _, exists := bashMap[rule.Pattern]; exists {
			delete(bashMap, rule.Pattern)
			removed++
			removedPatterns = append(removedPatterns, rule.Pattern)
		}
	}
	return
}
