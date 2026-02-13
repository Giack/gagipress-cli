package integration

import (
	"os"
	"testing"
)

// SkipIfNoSupabase skips test if Supabase credentials not available
func SkipIfNoSupabase(t *testing.T) {
	if os.Getenv("SUPABASE_URL") == "" {
		t.Skip("Skipping integration test: SUPABASE_URL not set")
	}
}

// SkipIfNoOpenAI skips test if OpenAI key not available
func SkipIfNoOpenAI(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}
}

// GetTestSupabaseURL returns Supabase URL for testing
func GetTestSupabaseURL() string {
	return os.Getenv("SUPABASE_URL")
}

// GetTestSupabaseKey returns Supabase key for testing
func GetTestSupabaseKey() string {
	return os.Getenv("SUPABASE_KEY")
}

// GetTestOpenAIKey returns OpenAI API key for testing
func GetTestOpenAIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}
