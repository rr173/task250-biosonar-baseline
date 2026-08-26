package store

import "encoding/json"

// jsonBytes marshals v to a JSON string, tolerating errors.
func jsonBytes(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// jsonUnmarshal parses a JSON string into v.
func jsonUnmarshal(s string, v interface{}) error {
	if s == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}
