package adhoc_test

import "fmt"

func requestShape(body map[string]any) map[string]string {
	out := map[string]string{}
	for k, v := range body {
		out[k] = fmt.Sprintf("%T", v)
	}
	return out
}

func toolNames(value any) []string {
	out := []string{}
	var walk func(any)
	walk = func(value any) {
		switch v := value.(type) {
		case []any:
			for _, x := range v {
				walk(x)
			}
		case map[string]any:
			if name, ok := v["name"].(string); ok {
				out = append(out, name)
			}
			for _, x := range v {
				walk(x)
			}
		}
	}
	walk(value)
	return out
}
