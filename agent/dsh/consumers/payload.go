package consumers

import (
	"bytes"
	"encoding/json"
)

// Payload readers are deliberately lenient: a row whose shape this build does
// not understand folds to a no-op instead of failing a whole session's stream.
// Envelope corruption is still caught by the log reader.

func decodeObject(raw []byte) map[string]json.RawMessage {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil
	}
	return fields
}

// objectField reads one object-valued field.
func objectField(fields map[string]json.RawMessage, key string) map[string]json.RawMessage {
	raw, ok := fields[key]
	if !ok {
		return nil
	}
	return decodeObject(raw)
}

// text reads one string-valued field.
func text(fields map[string]json.RawMessage, key string) string {
	raw, ok := fields[key]
	if !ok {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return value
}

// integer reads one integer-valued field.
func integer(fields map[string]json.RawMessage, key string) int64 {
	raw, ok := fields[key]
	if !ok {
		return 0
	}
	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		var float float64
		if json.Unmarshal(raw, &float) != nil {
			return 0
		}
		return int64(float)
	}
	return value
}

// boolean reads one boolean-valued field.
func boolean(fields map[string]json.RawMessage, key string) bool {
	raw, ok := fields[key]
	if !ok {
		return false
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	return value
}

// objectArray reads an array of objects.
func objectArray(fields map[string]json.RawMessage, key string) []map[string]json.RawMessage {
	raw, ok := fields[key]
	if !ok {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	objects := make([]map[string]json.RawMessage, 0, len(items))
	for _, item := range items {
		if object := decodeObject(item); object != nil {
			objects = append(objects, object)
		}
	}
	return objects
}

// argumentObject reads a tool's "arguments" field, which is a JSON document
// carried as a string.
func argumentObject(fields map[string]json.RawMessage, key string) map[string]json.RawMessage {
	encoded := text(fields, key)
	if encoded == "" {
		return nil
	}
	return decodeObject([]byte(encoded))
}
