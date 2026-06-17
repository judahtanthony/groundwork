// Package encoding holds the canonical value encodings that make stored data
// and Markdown exports deterministic (ADR 0020): RFC3339 UTC timestamps at
// second precision, and compact canonical JSON for the *_json columns.
package encoding

import (
	"encoding/json"
	"time"
)

// TimeLayout is the canonical timestamp layout: RFC3339, UTC, second precision.
const TimeLayout = "2006-01-02T15:04:05Z"

// FormatTime renders t in the canonical layout (UTC, truncated to the second).
func FormatTime(t time.Time) string {
	return t.UTC().Truncate(time.Second).Format(TimeLayout)
}

// ParseTime parses a canonical timestamp.
func ParseTime(s string) (time.Time, error) {
	return time.Parse(TimeLayout, s)
}

// Now returns the current time in the canonical layout.
func Now() string {
	return FormatTime(time.Now())
}

// JSON returns the canonical JSON encoding of v: compact, with object keys
// sorted (encoding/json sorts map keys) and array order preserved. It is used
// for the *_json columns so equal values serialize identically.
func JSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
