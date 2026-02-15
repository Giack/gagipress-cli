package models

import (
	"encoding/json"
	"time"
)

// Date represents a date without time component (YYYY-MM-DD format)
// Compatible with PostgreSQL DATE type
type Date struct {
	time.Time
}

const DateFormat = "2006-01-02"

// UnmarshalJSON implements json.Unmarshaler to handle YYYY-MM-DD format
func (d *Date) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		d.Time = time.Time{}
		return nil
	}

	// Remove quotes from JSON string
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// Handle empty string
	if str == "" {
		d.Time = time.Time{}
		return nil
	}

	// Parse YYYY-MM-DD format
	parsed, err := time.Parse(DateFormat, str)
	if err != nil {
		return err
	}

	// Normalize to UTC midnight
	d.Time = parsed.UTC()
	return nil
}

// MarshalJSON implements json.Marshaler to output YYYY-MM-DD format
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(DateFormat))
}

// String returns the date in YYYY-MM-DD format
func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format(DateFormat)
}
