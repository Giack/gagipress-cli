package models

import (
	"testing"
)

func TestBookInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		book    BookInput
		wantErr bool
	}{
		{
			name: "valid book",
			book: BookInput{
				Title:          "Test Book",
				Genre:          "children",
				TargetAudience: "3-5 years",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			book: BookInput{
				Genre:          "children",
				TargetAudience: "3-5 years",
			},
			wantErr: true,
		},
		{
			name: "missing genre",
			book: BookInput{
				Title:          "Test Book",
				TargetAudience: "3-5 years",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.book.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BookInput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
