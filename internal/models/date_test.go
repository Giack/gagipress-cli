package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDate    string // Expected date in YYYY-MM-DD format
		wantIsZero  bool
		wantErr     bool
	}{
		{
			name:       "valid date",
			input:      `"2026-01-20"`,
			wantDate:   "2026-01-20",
			wantIsZero: false,
			wantErr:    false,
		},
		{
			name:       "null value",
			input:      `null`,
			wantDate:   "",
			wantIsZero: true,
			wantErr:    false,
		},
		{
			name:       "empty string",
			input:      `""`,
			wantDate:   "",
			wantIsZero: true,
			wantErr:    false,
		},
		{
			name:       "leap year date",
			input:      `"2024-02-29"`,
			wantDate:   "2024-02-29",
			wantIsZero: false,
			wantErr:    false,
		},
		{
			name:       "invalid format RFC3339",
			input:      `"2026-01-20T15:04:05Z"`,
			wantDate:   "",
			wantIsZero: false,
			wantErr:    true,
		},
		{
			name:       "invalid format US",
			input:      `"01/20/2026"`,
			wantDate:   "",
			wantIsZero: false,
			wantErr:    true,
		},
		{
			name:       "invalid date",
			input:      `"2026-13-45"`,
			wantDate:   "",
			wantIsZero: false,
			wantErr:    true,
		},
		{
			name:       "malformed JSON",
			input:      `"incomplete`,
			wantDate:   "",
			wantIsZero: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := json.Unmarshal([]byte(tt.input), &d)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if d.IsZero() != tt.wantIsZero {
				t.Errorf("IsZero() = %v, want %v", d.IsZero(), tt.wantIsZero)
			}

			if !tt.wantIsZero {
				got := d.Format(DateFormat)
				if got != tt.wantDate {
					t.Errorf("Format() = %v, want %v", got, tt.wantDate)
				}
			}
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		date     Date
		wantJSON string
	}{
		{
			name:     "valid date",
			date:     Date{Time: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
			wantJSON: `"2026-01-20"`,
		},
		{
			name:     "zero date",
			date:     Date{Time: time.Time{}},
			wantJSON: `null`,
		},
		{
			name:     "leap year date",
			date:     Date{Time: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
			wantJSON: `"2024-02-29"`,
		},
		{
			name:     "date with time component ignored",
			date:     Date{Time: time.Date(2026, 1, 20, 15, 30, 45, 0, time.UTC)},
			wantJSON: `"2026-01-20"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.date)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}

			if string(got) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %v, want %v", string(got), tt.wantJSON)
			}
		})
	}
}

func TestDate_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		wantZero bool
	}{
		{
			name:     "normal date",
			dateStr:  "2026-01-20",
			wantZero: false,
		},
		{
			name:     "null date",
			dateStr:  "null",
			wantZero: true,
		},
		{
			name:     "leap year",
			dateStr:  "2024-02-29",
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with JSON string
			original := tt.dateStr
			if tt.dateStr != "null" {
				original = `"` + tt.dateStr + `"`
			}

			// Unmarshal to Date
			var d Date
			if err := json.Unmarshal([]byte(original), &d); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Marshal back to JSON
			got, err := json.Marshal(d)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Should match original
			if string(got) != original {
				t.Errorf("Round trip failed: got %v, want %v", string(got), original)
			}
		})
	}
}

func TestDate_PointerSemantics(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
		wantDate string
	}{
		{
			name:     "non-null pointer",
			input:    `{"date":"2026-01-20"}`,
			wantNil:  false,
			wantDate: "2026-01-20",
		},
		{
			name:     "null pointer",
			input:    `{"date":null}`,
			wantNil:  true, // JSON null means pointer is nil
			wantDate: "",
		},
		{
			name:     "omitted field",
			input:    `{}`,
			wantNil:  true, // omitempty means field is nil
			wantDate: "",
		},
	}

	type testStruct struct {
		Date *Date `json:"date,omitempty"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s testStruct
			if err := json.Unmarshal([]byte(tt.input), &s); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if (s.Date == nil) != tt.wantNil {
				t.Errorf("Date == nil: got %v, want %v", s.Date == nil, tt.wantNil)
			}

			if !tt.wantNil && tt.wantDate != "" {
				got := s.Date.Format(DateFormat)
				if got != tt.wantDate {
					t.Errorf("Format() = %v, want %v", got, tt.wantDate)
				}
			}
		})
	}
}

func TestDate_String(t *testing.T) {
	tests := []struct {
		name string
		date Date
		want string
	}{
		{
			name: "valid date",
			date: Date{Time: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
			want: "2026-01-20",
		},
		{
			name: "zero date",
			date: Date{Time: time.Time{}},
			want: "",
		},
		{
			name: "date with time component",
			date: Date{Time: time.Date(2026, 1, 20, 15, 30, 0, 0, time.UTC)},
			want: "2026-01-20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.date.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
