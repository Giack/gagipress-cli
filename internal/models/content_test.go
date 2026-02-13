package models

import (
	"testing"
)

func TestContentIdeaInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		idea    ContentIdeaInput
		wantErr bool
	}{
		{
			name: "valid idea",
			idea: ContentIdeaInput{
				Type:             "educational",
				BriefDescription: "Test idea",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			idea: ContentIdeaInput{
				Type:             "invalid",
				BriefDescription: "Test idea",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			idea: ContentIdeaInput{
				Type: "educational",
			},
			wantErr: true,
		},
		{
			name: "missing type",
			idea: ContentIdeaInput{
				BriefDescription: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.idea.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ContentIdeaInput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
