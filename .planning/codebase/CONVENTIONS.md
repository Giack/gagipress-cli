# Coding Conventions

**Analysis Date:** 2026-02-25

## Naming Patterns

**Files:**
- Package directories use lowercase: `internal/models/`, `internal/repository/`, `cmd/books/`
- Go source files use lowercase with underscores for test files: `book.go`, `book_test.go`
- Command files follow pattern: `cmd/books/add.go`, `cmd/books/list.go`, `cmd/calendar/plan.go`
- Test files paired with source: `internal/models/book.go` → `internal/models/book_test.go`

**Functions:**
- Exported functions use PascalCase: `NewBooksRepository()`, `GenerateIdeas()`, `ParseCSV()`
- Unexported functions use camelCase: `runAdd()`, `parseScriptFromResponse()`, `findColumn()`
- Constructor functions follow pattern: `func New<Type>(...) *<Type>` (e.g., `NewBooksRepository()`, `NewIdeaGenerator()`)
- Error handling functions: `New()`, `Wrap()`, `IsType()` for error creation and checking

**Variables:**
- Local variables use camelCase: `reader`, `bookInput`, `apiKey`, `bookTitle`
- Constants use SCREAMING_SNAKE_CASE: `ErrorTypeValidation`, `ErrorTypeAPI`, `DateFormat`
- Type enums defined as constants: `ErrorType` string constants like `ErrorTypeValidation`
- Short loop/index variables accepted: `i`, `j`, `tt` (for table-driven tests)

**Types:**
- Struct types use PascalCase: `Book`, `ContentIdea`, `ContentScript`, `BookInput`
- Interface types use PascalCase (but not in codebase heavily): N/A
- Type aliases use PascalCase: `ErrorType`, `Date`

**Config structs:**
- All config structs have BOTH `mapstructure` and `yaml` tags (critical requirement):
  ```go
  type SupabaseConfig struct {
      URL        string `mapstructure:"url" yaml:"url"`
      AnonKey    string `mapstructure:"anon_key" yaml:"anon_key"`
      ServiceKey string `mapstructure:"service_key" yaml:"service_key"`
  }
  ```
- Reason: `mapstructure` for Viper deserialization, `yaml` for serialization (WriteConfigAs)

## Code Style

**Formatting:**
- Tool: `go fmt` (built-in Go formatter)
- Run before commits: `make fmt`
- Makefile target available for consistency

**Linting:**
- Tool: `go vet` for static analysis
- Run before commits: `make vet`
- Enforces safe/idiomatic code patterns

**Indentation:**
- Tabs (Go standard, enforced by gofmt)
- No custom indentation rules observed

**Brace Style:**
- Opening braces on same line: `func foo() {`
- K&R style throughout codebase

## Import Organization

**Order (observed in codebase):**
1. Standard library imports: `"fmt"`, `"time"`, `"context"`, `"encoding/json"`
2. External packages: `"github.com/spf13/cobra"`, `"github.com/spf13/viper"`
3. Internal packages: `"github.com/gagipress/gagipress-cli/internal/..."`

Example from `cmd/books/add.go`:
```go
import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)
```

**Path Aliases:**
- No aliases used in codebase
- Full import paths: `github.com/gagipress/gagipress-cli/internal/...`

## Error Handling

**Patterns:**
- Custom `AppError` type in `internal/errors/errors.go` with `Type` field:
  ```go
  type AppError struct {
      Type    ErrorType  // Category: validation, api, database, not_found, network
      Message string     // Descriptive message
      Err     error      // Wrapped underlying error
  }
  ```
- Error creation functions:
  - `errors.New(ErrorType, message)` - Create new error
  - `errors.Wrap(err, ErrorType, message)` - Wrap existing error with context
  - `errors.IsType(err, ErrorType)` - Check error type

- Commands wrap errors with context:
  ```go
  if err != nil {
      return fmt.Errorf("failed to load config: %w", err)
  }
  ```

**Retry Logic:**
- Exponential backoff in `internal/errors/retry.go`
- Used by OpenAI client for API failures:
  ```go
  retryErr := errors.Retry(ctx, errors.DefaultRetryConfig(), func() error {
      responseText, err = g.openaiClient.GenerateText(prompt, 0.8)
      if err != nil {
          return errors.Wrap(err, errors.ErrorTypeAPI, "OpenAI API call failed")
      }
      return nil
  })
  ```
- Does NOT retry on validation errors (only server errors)
- Respects context cancellation

**Repository Pattern:**
- All database operations through `internal/repository/*` packages
- Each repository handles single domain: `BooksRepository`, `ContentRepository`, `CalendarRepository`
- Returns domain models, not raw API responses

## Logging

**Framework:** `fmt` package (standard library)

**Patterns:**
- Status messages with emojis: `fmt.Println("📚 Add New Book")`, `fmt.Println("💾 Saving book...")`
- Error messages: `fmt.Printf("⚠️  Invalid date format, skipping publication date\n")`
- Success feedback: `fmt.Println("✅ Book added successfully!")`
- Info messages with formatting: `fmt.Printf("   ID: %s\n", book.ID)`

**When to Log:**
- User-facing progress indicators in commands
- Not used for debug logging in internal packages
- Errors returned instead of logged

## Comments

**When to Comment:**
- Public (exported) functions/types have doc comments
- Complex logic explained inline (rare in this codebase)
- NOT used for obvious code (e.g., "increment counter")

**JSDoc/GoDoc Style:**
- Comments start with function name for exported functions:
  ```go
  // Validate validates book input
  func (b *BookInput) Validate() error

  // GenerateIdeas generates content ideas for a book
  func (g *IdeaGenerator) GenerateIdeas(...)
  ```
- Type documentation above struct definition:
  ```go
  // Book represents a book in the catalog
  type Book struct

  // BookInput represents input for creating/updating a book
  type BookInput struct
  ```
- Brief, one-line descriptions (no detailed explanations)

## Function Design

**Size:** Functions are typically 10-50 lines; longer functions in repositories (HTTP calls with parsing)

**Parameters:**
- Receiver methods for domain models: `func (b *BookInput) Validate() error`
- Constructor receives config: `func NewBooksRepository(cfg *SupabaseConfig) *BooksRepository`
- No excessive parameters (max 3-4 observed)
- Receiver pointers for modification, values for read-only

**Return Values:**
- Error as last return: `func (r *BooksRepository) GetAll() ([]Book, error)`
- Named returns avoided (single error returns only)
- Multiple values packed in structs when complex:
  ```go
  type ChatCompletionResponse struct {
      Message string
      Tokens  int
  }
  ```

**Input Validation:**
- Validation methods on input structs: `BookInput.Validate()`, `ContentIdea.Validate()`
- Validate early in command before operations
- Repository methods assume inputs are valid (validation at boundary)

## Module Design

**Exports:**
- Only types/functions that need to be public are exported (PascalCase)
- Helper functions lowercase/unexported: `findColumn()`, `parseIdeasFromResponse()`
- Constants for error types exported: `ErrorTypeValidation`, `ErrorTypeAPI`

**Barrel Files:**
- No barrel files (index.go) used in this codebase
- Each package directly imports what it needs

**Package Structure:**
- `internal/` packages never imported from outside `internal/`
- Clear separation: `cmd/` for CLI commands, `internal/` for business logic
- Commands thin wrappers around `internal/` packages

## Struct Tags

**JSON Tags:**
- Use for API serialization: `json:"id"`, `json:"title"`
- Support omitempty for optional fields: `json:"target_audience,omitempty"`

**Config Tags (CRITICAL):**
- MUST include both tags (mapstructure + yaml):
  ```go
  type SupabaseConfig struct {
      URL string `mapstructure:"url" yaml:"url"`
  }
  ```
- Use snake_case in tags: `anon_key`, `service_key` (not camelCase)
- Reason: mapstructure for reading, yaml for writing

## Testing Naming

**Test Functions:**
- Pattern: `Test<FunctionName>_<Scenario>`
- Examples:
  - `TestGetBookByIDPrefix_ValidPrefix_FindsUniqueBook()`
  - `TestBookInput_Validate()`
  - `TestKDPParser_RoyaltyParsing()`

**Table-Driven Tests:**
- Use `tests` slice of anonymous structs
- Variable `tt` for loop iteration: `for _, tt := range tests`
- Common fields: `name`, `input`, `wantErr`, `wantCount`, `wantJSON`

## Directory Structure

- `cmd/` - Cobra commands, thin CLI wrappers
- `internal/` - Business logic, not importable from outside
  - `models/` - Domain types with validation
  - `repository/` - Database/API access (Supabase REST)
  - `config/` - Configuration loading/saving
  - `errors/` - Error types and retry logic
  - `ai/` - OpenAI and Gemini clients
  - `generator/` - Content generation (ideas, scripts)
  - `scheduler/` - Calendar planning algorithms
  - `parser/` - CSV parsing (KDP reports)
  - `prompts/` - AI prompt templates
  - `social/` - Social media API clients
  - `ui/` - CLI UI components (tables, formatters)
- `test/integration/` - Integration tests (require credentials)

---

*Convention analysis: 2026-02-25*
