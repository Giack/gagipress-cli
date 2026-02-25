# Testing Patterns

**Analysis Date:** 2026-02-25

## Test Framework

**Runner:**
- Go's built-in `testing` package (no third-party test framework)
- Config: `go test ./...` (all tests)
- Version: Go 1.24.0+

**Assertion Library:**
- Manual assertions using `t.Error()`, `t.Errorf()`, `t.Fatal()`, `t.Fatalf()`
- No external assertion library (e.g., stretchr/testify)
- Simple string comparisons and error checks

**Run Commands:**
```bash
make test              # Run all tests
make test-integration  # Run integration tests only
make test-coverage     # Run tests with coverage report
```

**Test Execution:**
```bash
go test ./...                    # Run all tests
go test ./internal/parser/...    # Run specific package
go test -run TestName            # Run specific test
go test -v                       # Verbose output with individual test names
```

## Test File Organization

**Location:**
- Unit tests co-located with source code (same package)
- Pattern: `file.go` paired with `file_test.go` in same directory
- Examples:
  - `internal/models/book.go` → `internal/models/book_test.go`
  - `internal/parser/kdp.go` → `internal/parser/kdp_test.go`
  - `internal/errors/errors.go` → `internal/errors/errors_test.go`

**Naming:**
- Test files end with `_test.go`
- Test functions: `Test<Type>_<Scenario>` or `Test<Function>_<Scenario>`

**Structure:**
```
internal/
├── models/
│   ├── book.go
│   ├── book_test.go
│   ├── date.go
│   ├── date_test.go
│   └── content.go
├── parser/
│   ├── kdp.go
│   └── kdp_test.go
└── repository/
    ├── books.go
    ├── books_test.go
    ├── content.go
    └── content_test.go
```

**Integration Tests:**
- Located in `test/integration/` directory
- Examples:
  - `test/integration/workflow_test.go`
  - `test/integration/approve_reject_test.go`
  - `test/integration/setup_test.go`

**Test Utilities:**
- Helper functions in `internal/testutil/`
- Shared test utilities and fixtures

## Test Structure

**Suite Organization:**
```go
// Table-driven test pattern (most common)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.book.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Patterns:**

1. **Table-Driven Tests** (primary pattern):
   - Used extensively: `parser/kdp_test.go`, `models/date_test.go`, `config/config_test.go`
   - Struct fields: `name`, `input`/test data, `want<Result>`, `wantErr`
   - Loop with `t.Run(tt.name, ...)` for subtest naming

2. **Setup/Teardown:**
   - `t.TempDir()` for temporary test directories
   - HTTP mock servers via `httptest.NewServer()`
   - Example from `repository/books_test.go`:
     ```go
     func newTestBooksRepo(handler http.HandlerFunc) *BooksRepository {
         server := httptest.NewServer(handler)
         cfg := &config.SupabaseConfig{
             URL:     server.URL,
             AnonKey: "test-key",
         }
         return NewBooksRepository(cfg)
     }
     ```

3. **Assertion Pattern:**
   - Direct comparison: `if result != expected { t.Errorf(...) }`
   - Error checking: `if (err != nil) != wantErr { ... }`
   - String contains: Custom helper or substring check
   - Example:
     ```go
     if len(rows) != tt.wantCount {
         t.Errorf("ParseCSV() got %d rows, want %d", len(rows), tt.wantCount)
     }
     ```

## Mocking

**Framework:** Manual mocking using `httptest` for HTTP endpoints

**Patterns:**

1. **HTTP Server Mocking** (most common):
   ```go
   handler := func(w http.ResponseWriter, r *http.Request) {
       // Verify request
       if r.Method != "POST" {
           t.Errorf("expected POST, got %s", r.Method)
       }
       // Send response
       w.WriteHeader(http.StatusOK)
       json.NewEncoder(w).Encode(responseData)
   }
   server := httptest.NewServer(handler)
   defer server.Close()
   ```
   - Used in `repository/books_test.go`, `repository/content_test.go`

2. **Struct Initialization:**
   - Create test structs with specific values
   - Example from `models/date_test.go`:
     ```go
     date: Date{Time: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)}
     ```

**What to Mock:**
- HTTP/network calls (via httptest)
- External API responses
- Repository layer in service tests

**What NOT to Mock:**
- Standard library functions (time, json, etc.)
- Internal calculations and logic
- Validation logic (test directly)
- CSV parsing (test with real data strings)

## Fixtures and Factories

**Test Data:**
- Inline test data in table-driven tests
- Example from `parser/kdp_test.go`:
  ```go
  csvData: `Title,ASIN,Date,Units Sold,Royalty
Test Book,B0ABC123,2024-01-15,5,$3.50`
  ```
- Multi-line strings for CSV/complex structures

**Factories:**
- Constructor functions used for test repos:
  ```go
  func newTestBooksRepo(handler http.HandlerFunc) *BooksRepository {
      server := httptest.NewServer(handler)
      return NewBooksRepository(&config.SupabaseConfig{
          URL:     server.URL,
          AnonKey: "test-key",
      })
  }
  ```

**Location:**
- Inline in test files alongside tests
- No separate fixtures directory

## Coverage

**Requirements:** Not enforced (no CI/linting rule)

**Current Coverage (as of 2026-02-25):**
- `errors` package: 96.3% (critical retry logic)
- `parser` package: 96.9% (battle-tested CSV parsing)
- `models` package: 53.6% (validation and serialization)
- `scheduler` package: 51.7% (planning algorithms)
- `ui` package: 32.7% (formatting utilities)
- `repository` package: 21.9% (mostly HTTP mocking heavy)
- `ai`, `generator`, `social`, `prompts`: 0% (stubs/integration only)
- Commands (`cmd/*`): 0% (manual testing only)
- **Overall: 11.5%**

**View Coverage:**
```bash
make test-coverage  # Generates HTML report
open coverage.html  # View in browser
```

**Coverage Script:** `scripts/test-coverage.sh`
- Runs tests with `-coverprofile=coverage.out`
- Generates HTML report: `go tool cover -html=coverage.out`

## Test Types

**Unit Tests:**
- Scope: Individual functions/methods
- Approach: Table-driven with mock HTTP servers
- Examples:
  - `models/book_test.go` - Validation logic
  - `parser/kdp_test.go` - CSV parsing edge cases (commas in currency, date formats)
  - `errors/retry_test.go` - Exponential backoff behavior
  - `ui/formatters_test.go` - Number/date formatting
  - `models/date_test.go` - JSON marshaling/unmarshaling with null handling

**Integration Tests:**
- Location: `test/integration/`
- Require: Supabase credentials (skipped without)
- Examples:
  - `test/integration/workflow_test.go` - Full content generation → approval → scheduling
  - `test/integration/approve_reject_test.go` - Idea approval workflow
  - `test/integration/setup_test.go` - Database connection and schema

**E2E Tests:**
- Not formally structured
- Manual testing via CLI commands
- Example: `gagipress test gemini "Hello" --headless=false`

## Common Patterns

**Async Testing:**
Not used (Go's synchronous testing model sufficient)

**Error Testing:**
```go
// Pattern 1: Check error existence
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}

// Pattern 2: Table-driven with wantErr
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        wantErr bool
    }{
        {"success", false},
        {"failure", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := someFunc()
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

// Pattern 3: Error type checking
if !strings.Contains(err.Error(), "expected message") {
    t.Errorf("expected error message, got: %v", err)
}
```

**JSON Serialization Testing:**
```go
// Round-trip test: unmarshal → marshal → compare
func TestDate_RoundTrip(t *testing.T) {
    original := `"2026-01-20"`
    var d Date
    json.Unmarshal([]byte(original), &d)
    got, _ := json.Marshal(d)
    if string(got) != original {
        t.Errorf("round trip failed: got %v, want %v", string(got), original)
    }
}
```

**Repository Testing:**
```go
// HTTP mock server pattern
func TestGetBookByIDPrefix_ValidPrefix(t *testing.T) {
    handler := func(w http.ResponseWriter, r *http.Request) {
        // Assert request correctness
        if r.Method != "POST" {
            t.Errorf("expected POST, got %s", r.Method)
        }
        // Return mock response
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode([]models.Book{testBook})
    }

    server := httptest.NewServer(handler)
    defer server.Close()

    repo := NewBooksRepository(&config.SupabaseConfig{
        URL:     server.URL,
        AnonKey: "test-key",
    })

    result, err := repo.GetBookByIDPrefix("abcdef12")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.ID != testBook.ID {
        t.Errorf("expected ID %s, got %s", testBook.ID, result.ID)
    }
}
```

## Test Gaps (from Coverage Analysis)

**High Priority (0% coverage):**
- `internal/ai/*` - OpenAI and Gemini client integration
- `internal/generator/*` - Content idea/script generation
- `internal/social/*` - Instagram, TikTok, Blotato API clients
- `cmd/*` - All CLI commands (manual testing only)

**Medium Priority (<50% coverage):**
- `internal/repository/*` - Database operations (21.9%)
  - Gap: Create, Update, Delete operations for most models
  - Gap: Error handling for API failures
- `internal/ui/*` - Table rendering, terminal utilities (32.7%)
  - Gap: Complex table layout with word wrapping

**Why Low Coverage:**
- AI providers (OpenAI, Gemini) require live API keys
- Social media APIs require OAuth tokens
- Database operations tested via integration tests
- CLI commands require user interaction (better for manual/e2e testing)

## Testing Best Practices (from CLAUDE.md)

**Coverage Targets:**
- Critical code (parsers, error handling): 80%+
- Business logic (scheduler, models): 40-60%
- Commands: Manual testing (not unit tested)

**Testing Strategy:**
- Mock at repository layer, not HTTP layer
- Integration tests skip gracefully without credentials
- Use `internal/testutil/helpers.go` for common assertions

**Known Tested Edge Cases:**
- CSV parsing: Commas in currency fields (`"$1,234.56"`)
- Date parsing: Multiple formats (ISO, US, EU, slash)
- Retry logic: Exponential backoff, context cancellation, validation error handling
- Config serialization: Both mapstructure and yaml tags
- Date JSON: Null vs empty, pointer semantics, round-trip serialization

---

*Testing analysis: 2026-02-25*
