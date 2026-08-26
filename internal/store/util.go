package store

import (
	"encoding/json"
	"fmt"
	"time"
)

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

// formatTimestamp renders a timestamp as a fixed-width RFC3339 string whose
// fractional seconds are always present and zero-padded to nine digits
// (e.g. "2026-01-02T15:04:05.000000000Z"). Because every timestamp shares the
// same fixed structure, lexicographic ordering of these strings equals
// chronological ordering, so SQL's MAX(ts) returns the genuinely latest value
// even when several pings arrive within the same calendar second with
// differing sub-second offsets. The result remains parseable by RFC3339Nano.
func formatTimestamp(t time.Time) string {
	return fmt.Sprintf("%s.%09dZ",
		t.UTC().Format("2006-01-02T15:04:05"),
		t.UTC().Nanosecond())
}
