# Scripts Finalize & Export Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `scripts list`, `scripts finalize`, and `calendar export` commands to close the workflow gap between script generation and content publication.

**Architecture:** Three new CLI commands backed by new repository methods. The `finalize` command composes platform-ready captions, exports to markdown files, and transitions statuses. A new SQL migration adds a `find_script_by_prefix` function. The scheduler is updated to use only finalized scripts.

**Tech Stack:** Go 1.24+, Cobra CLI, Supabase REST API, lipgloss (terminal styling)

---

### Task 1: Add `Status` field to `ContentScript` model

**Files:**
- Modify: `internal/models/content.go:49-59`

**Step 1: Add Status field to ContentScript struct**

Add `Status` between `EstimatedDuration` and `CreatedAt` in the `ContentScript` struct:

```go
// ContentScript represents a generated script
type ContentScript struct {
	ID                string    `json:"id"`
	IdeaID            string    `json:"idea_id"`
	Hook              string    `json:"hook"`
	FullScript        string    `json:"full_script"`
	CTA               string    `json:"cta"`
	Hashtags          []string  `json:"hashtags,omitempty"`
	EstimatedDuration int       `json:"estimated_duration"` // seconds
	Status            string    `json:"status"`             // draft, finalized, used
	CreatedAt         time.Time `json:"created_at"`
}
```

**Step 2: Verify it compiles**

Run: `make build`
Expected: Build succeeds (Status is read-only, no write changes needed)

**Step 3: Commit**

```bash
git add internal/models/content.go
git commit -m "feat(models): add Status field to ContentScript struct"
```

---

### Task 2: Add `FormatStatus` support for new statuses

**Files:**
- Modify: `internal/ui/styles.go:24-46`

**Step 1: Add new badge styles and update FormatStatus**

Add badges for `draft`, `finalized`, `scripted`, `pending_approval` in `styles.go`:

```go
// Status badge styles
var (
	BadgePending        = lipgloss.NewStyle().Foreground(ColorWarning).Render("pending")
	BadgeApproved       = lipgloss.NewStyle().Foreground(ColorSuccess).Render("approved")
	BadgeRejected       = lipgloss.NewStyle().Foreground(ColorError).Render("rejected")
	BadgeDraft          = lipgloss.NewStyle().Foreground(ColorMuted).Render("draft")
	BadgeFinalized      = lipgloss.NewStyle().Foreground(ColorPrimary).Render("finalized")
	BadgeScripted       = lipgloss.NewStyle().Foreground(ColorPrimary).Render("scripted")
	BadgePendingApproval = lipgloss.NewStyle().Foreground(ColorWarning).Render("pending_approval")
)

// FormatStatus returns a colored status badge
func FormatStatus(status string) string {
	if !IsColorTerminal() {
		return status // Graceful degradation
	}

	switch status {
	case "pending":
		return BadgePending
	case "approved":
		return BadgeApproved
	case "rejected":
		return BadgeRejected
	case "draft":
		return BadgeDraft
	case "finalized":
		return BadgeFinalized
	case "scripted":
		return BadgeScripted
	case "pending_approval":
		return BadgePendingApproval
	default:
		return status
	}
}
```

**Step 2: Verify it compiles**

Run: `make build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/ui/styles.go
git commit -m "feat(ui): add status badges for draft, finalized, scripted"
```

---

### Task 3: Add repository methods — `UpdateScriptStatus`, `GetScripts` status filter, `GetScriptByID`

**Files:**
- Modify: `internal/repository/content.go:282-324`
- Test: `internal/repository/content_test.go`

**Step 1: Write failing tests for UpdateScriptStatus**

Add to `internal/repository/content_test.go`:

```go
func TestUpdateScriptStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.String(), "content_scripts") {
			t.Errorf("expected content_scripts endpoint, got: %s", r.URL.String())
		}
		if !strings.Contains(r.URL.String(), "id=eq.test-id") {
			t.Errorf("expected id filter, got: %s", r.URL.String())
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["status"] != "finalized" {
			t.Errorf("expected status finalized, got: %s", body["status"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})
	err := repo.UpdateScriptStatus("test-id", "finalized")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetScripts_WithStatusFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "status=eq.draft") {
			t.Errorf("expected status filter, got: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentScript{})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})
	_, err := repo.GetScripts("draft", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetScriptByID_Success(t *testing.T) {
	scriptID := "abcdef12-3456-7890-abcd-ef1234567890"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "id=eq."+scriptID) {
			t.Errorf("expected id filter, got: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentScript{
			{ID: scriptID, Hook: "test", Status: "draft"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})
	script, err := repo.GetScriptByID(scriptID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script.ID != scriptID {
		t.Errorf("expected ID %s, got %s", scriptID, script.ID)
	}
}

func TestGetScriptByID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentScript{})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})
	_, err := repo.GetScriptByID("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %s", err.Error())
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `make test`
Expected: FAIL — `UpdateScriptStatus`, `GetScriptByID` undefined

**Step 3: Implement the methods**

Add to `internal/repository/content.go`:

```go
// UpdateScriptStatus updates the status of a content script
func (r *ContentRepository) UpdateScriptStatus(id string, status string) error {
	url := fmt.Sprintf("%s/rest/v1/content_scripts?id=eq.%s", r.config.URL, id)

	data := map[string]string{"status": status}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update script: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetScriptByID retrieves a single content script by ID
func (r *ContentRepository) GetScriptByID(id string) (*models.ContentScript, error) {
	url := fmt.Sprintf("%s/rest/v1/content_scripts?id=eq.%s&select=*", r.config.URL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get script: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get script: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var scripts []models.ContentScript
	if err := json.Unmarshal(body, &scripts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(scripts) == 0 {
		return nil, fmt.Errorf("script not found: %s", id)
	}

	return &scripts[0], nil
}
```

Also modify `GetScripts` to accept a `status` parameter. Change the signature from `GetScripts(limit int)` to `GetScripts(status string, limit int)`:

```go
// GetScripts retrieves content scripts
func (r *ContentRepository) GetScripts(status string, limit int) ([]models.ContentScript, error) {
	url := fmt.Sprintf("%s/rest/v1/content_scripts?select=*&order=created_at.desc", r.config.URL)

	if status != "" {
		url += fmt.Sprintf("&status=eq.%s", status)
	}
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	// ... rest unchanged
```

**Step 4: Update all callers of GetScripts to pass status**

There are two callers:
- `internal/scheduler/planner.go:29` — change `GetScripts(0)` to `GetScripts("", 0)` (we'll change to `"finalized"` in Task 8)
- No other callers

**Step 5: Run tests to verify they pass**

Run: `make test`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/repository/content.go internal/repository/content_test.go internal/scheduler/planner.go
git commit -m "feat(repo): add UpdateScriptStatus, GetScriptByID, status filter on GetScripts"
```

---

### Task 4: Add SQL migration for `find_script_by_prefix`

**Files:**
- Create: `migrations/004_find_script_by_prefix.sql`
- Create: `supabase/migrations/004_find_script_by_prefix.sql`

**Step 1: Create migration file**

Create `migrations/004_find_script_by_prefix.sql`:

```sql
-- Migration: Add UUID prefix matching function for content_scripts
-- Same pattern as find_idea_by_prefix (migration 003)

CREATE OR REPLACE FUNCTION find_script_by_prefix(prefix_pattern TEXT)
RETURNS SETOF content_scripts
LANGUAGE sql
STABLE
AS $$
  SELECT *
  FROM content_scripts
  WHERE id::text LIKE prefix_pattern || '%';
$$;

GRANT EXECUTE ON FUNCTION find_script_by_prefix(TEXT) TO anon, authenticated;

COMMENT ON FUNCTION find_script_by_prefix IS 'Find content scripts by UUID prefix (case-sensitive). Example: find_script_by_prefix(''abcd1234'')';
```

**Step 2: Copy to supabase/migrations/**

Copy the same file to `supabase/migrations/004_find_script_by_prefix.sql` (Supabase CLI expects migrations here).

**Step 3: Commit**

```bash
git add migrations/004_find_script_by_prefix.sql supabase/migrations/004_find_script_by_prefix.sql
git commit -m "feat(db): add find_script_by_prefix SQL function"
```

---

### Task 5: Add `GetScriptByIDPrefix` repository method

**Files:**
- Modify: `internal/repository/content.go`
- Test: `internal/repository/content_test.go`

**Step 1: Write failing tests**

Add to `internal/repository/content_test.go`:

```go
func TestGetScriptByIDPrefix_PrefixTooShort(t *testing.T) {
	repo := NewContentRepository(&config.SupabaseConfig{URL: "http://localhost", AnonKey: "test"})

	tests := []struct {
		name   string
		prefix string
	}{
		{"empty", ""},
		{"1 char", "a"},
		{"5 chars", "abcde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.GetScriptByIDPrefix(tt.prefix)
			if err == nil {
				t.Fatal("expected error for short prefix, got nil")
			}
			if got := err.Error(); !contains(got, "prefix too short") {
				t.Errorf("expected 'prefix too short' error, got: %s", got)
			}
		})
	}
}

func TestGetScriptByIDPrefix_SingleMatch(t *testing.T) {
	fullID := "abcdef12-3456-7890-abcd-ef1234567890"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rpc/find_script_by_prefix") {
			t.Errorf("expected RPC endpoint, got path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentScript{
			{ID: fullID, Hook: "test hook", Status: "draft"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	script, err := repo.GetScriptByIDPrefix("abcdef12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script.ID != fullID {
		t.Errorf("expected ID %s, got %s", fullID, script.ID)
	}
}

func TestGetScriptByIDPrefix_NoMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentScript{})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	_, err := repo.GetScriptByIDPrefix("abcdef")
	if err == nil {
		t.Fatal("expected error for no matches, got nil")
	}
	if got := err.Error(); !contains(got, "no script found") {
		t.Errorf("expected 'no script found' error, got: %s", got)
	}
}

func TestGetScriptByIDPrefix_MultipleMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentScript{
			{ID: "abcdef12-1111-1111-1111-111111111111", Status: "draft"},
			{ID: "abcdef12-2222-2222-2222-222222222222", Status: "draft"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	_, err := repo.GetScriptByIDPrefix("abcdef12")
	if err == nil {
		t.Fatal("expected error for multiple matches, got nil")
	}
	if got := err.Error(); !contains(got, "ambiguous") {
		t.Errorf("expected 'ambiguous' error, got: %s", got)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `make test`
Expected: FAIL — `GetScriptByIDPrefix` undefined

**Step 3: Implement GetScriptByIDPrefix**

Add to `internal/repository/content.go` (same pattern as `GetIdeaByIDPrefix`):

```go
// GetScriptByIDPrefix finds a content script by UUID prefix (minimum 6 characters).
func (r *ContentRepository) GetScriptByIDPrefix(prefix string) (*models.ContentScript, error) {
	if len(prefix) < 6 {
		return nil, fmt.Errorf("prefix too short: must be at least 6 characters, got %d", len(prefix))
	}

	requestURL := fmt.Sprintf("%s/rest/v1/rpc/find_script_by_prefix", r.config.URL)

	reqBody := map[string]string{"prefix_pattern": prefix}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get script by prefix: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get script by prefix: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var scripts []models.ContentScript
	if err := json.Unmarshal(body, &scripts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	switch len(scripts) {
	case 0:
		return nil, fmt.Errorf("no script found with ID prefix %q", prefix)
	case 1:
		return &scripts[0], nil
	default:
		ids := make([]string, len(scripts))
		for i, s := range scripts {
			ids[i] = s.ID
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches %d scripts: %s", prefix, len(scripts), strings.Join(ids, ", "))
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `make test`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/repository/content.go internal/repository/content_test.go
git commit -m "feat(repo): add GetScriptByIDPrefix with prefix matching"
```

---

### Task 6: Create `cmd/scripts/` command group with `list` command

**Files:**
- Create: `cmd/scripts/scripts.go`
- Create: `cmd/scripts/list.go`
- Modify: `cmd/root.go:7-59`

**Step 1: Create the scripts command group**

Create `cmd/scripts/scripts.go`:

```go
package scripts

import (
	"github.com/spf13/cobra"
)

// ScriptsCmd represents the scripts command group
var ScriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Manage generated scripts",
	Long: `Manage generated content scripts:
  - List all scripts with filters
  - Finalize scripts for publishing`,
}
```

**Step 2: Create the list command**

Create `cmd/scripts/list.go`:

```go
package scripts

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	listStatusFilter string
	listLimit        int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List generated scripts",
	Long:  `Display all generated scripts with optional status filter.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&listStatusFilter, "status", "", "Filter by status (draft, finalized)")
	listCmd.Flags().IntVar(&listLimit, "limit", 50, "Maximum number of scripts to show")

	ScriptsCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repo := repository.NewContentRepository(&cfg.Supabase)
	scripts, err := repo.GetScripts(listStatusFilter, listLimit)
	if err != nil {
		return fmt.Errorf("failed to get scripts: %w", err)
	}

	if len(scripts) == 0 {
		fmt.Println("No scripts found. Generate some with 'gagipress generate script <idea-id>'")
		return nil
	}

	rows := make([][]string, len(scripts))
	for i, script := range scripts {
		// Platform hint from duration
		platformHint := "tiktok"
		if script.EstimatedDuration > 60 {
			platformHint = "instagram"
		}

		// Truncate hook for display
		hook := script.Hook
		if len(hook) > 50 {
			hook = hook[:47] + "..."
		}

		rows[i] = []string{
			script.ID,
			hook,
			platformHint,
			fmt.Sprintf("%ds", script.EstimatedDuration),
			ui.FormatStatus(script.Status),
			ui.FormatDate(script.CreatedAt),
		}
	}

	table := ui.RenderTable(ui.TableConfig{
		Headers:  []string{"ID", "Hook", "Platform", "Duration", "Status", "Created"},
		Rows:     rows,
		MaxWidth: ui.GetTerminalWidth(),
	})

	fmt.Println(ui.StyleHeader.Render("📝 Content Scripts"))
	fmt.Println(table)
	fmt.Printf("\nTotal scripts: %d\n", len(scripts))

	// Count drafts
	draftCount := 0
	for _, s := range scripts {
		if s.Status == "draft" {
			draftCount++
		}
	}
	if draftCount > 0 {
		fmt.Printf("\n📝 %d draft scripts ready to finalize\n", draftCount)
		fmt.Println("   Use 'gagipress scripts finalize <id>' to finalize")
	}

	return nil
}
```

**Step 3: Register in root.go**

Add import and command registration in `cmd/root.go`:

Add import: `"github.com/gagipress/gagipress-cli/cmd/scripts"`

Add line after `rootCmd.AddCommand(stats.StatsCmd)`:
```go
rootCmd.AddCommand(scripts.ScriptsCmd)
```

**Step 4: Verify it compiles**

Run: `make build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add cmd/scripts/scripts.go cmd/scripts/list.go cmd/root.go
git commit -m "feat(scripts): add scripts command group with list command"
```

---

### Task 7: Create `scripts finalize` command

**Files:**
- Create: `cmd/scripts/finalize.go`

**Step 1: Create the finalize command**

Create `cmd/scripts/finalize.go`:

```go
package scripts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var finalizeCmd = &cobra.Command{
	Use:   "finalize [script-id]",
	Short: "Finalize a script for publishing",
	Long: `Finalize a draft script, making it ready for publishing.

This will:
  - Compose a platform-ready caption (hook + CTA + hashtags)
  - Show a preview in the terminal
  - Save an export file to exports/
  - Update script status to "finalized"
  - Update idea status to "finalized"`,
	Args: cobra.ExactArgs(1),
	RunE: runFinalize,
}

func init() {
	ScriptsCmd.AddCommand(finalizeCmd)
}

func runFinalize(cmd *cobra.Command, args []string) error {
	scriptID := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("🎬 Finalize Script"))
	fmt.Println()

	repo := repository.NewContentRepository(&cfg.Supabase)

	// Resolve script ID
	fmt.Print(ui.StyleMuted.Render("Resolving script ID... "))
	script, err := repo.GetScriptByIDPrefix(scriptID)
	if err != nil {
		fmt.Println(ui.StyleError.Render("✗ FAILED"))
		return fmt.Errorf("failed to resolve script ID: %w", err)
	}
	scriptID = script.ID
	fmt.Println(ui.StyleSuccess.Render("✓ " + scriptID))

	// Check status
	if script.Status != "draft" {
		return fmt.Errorf("script must be in draft status (current: %s)", script.Status)
	}

	// Platform hint
	platform := "tiktok"
	if script.EstimatedDuration > 60 {
		platform = "instagram"
	}

	// Compose caption
	hashtagStr := strings.Join(script.Hashtags, " ")
	caption := script.Hook + "\n\n" + script.CTA + "\n\n" + hashtagStr

	// Display preview
	fmt.Println()
	fmt.Println(ui.StyleHeader.Render("📋 Caption Preview"))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(caption)
	fmt.Println(strings.Repeat("─", 60))

	fmt.Println()
	fmt.Println(ui.StyleHeader.Render("📄 Full Script"))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(script.FullScript)
	fmt.Println(strings.Repeat("─", 60))

	fmt.Printf("\n⏱️  Duration: %ds | 🎯 Platform: %s\n", script.EstimatedDuration, platform)

	// Save export file
	fmt.Print("\n💾 Saving export file... ")
	exportPath, err := saveExportFile(script, platform, caption)
	if err != nil {
		fmt.Println(ui.StyleError.Render("✗ FAILED"))
		return fmt.Errorf("failed to save export: %w", err)
	}
	fmt.Println(ui.StyleSuccess.Render("✓ " + exportPath))

	// Update script status
	fmt.Print(ui.StyleMuted.Render("Updating script status... "))
	if err := repo.UpdateScriptStatus(scriptID, "finalized"); err != nil {
		fmt.Println(ui.StyleError.Render("✗ FAILED"))
		return fmt.Errorf("failed to update script status: %w", err)
	}
	fmt.Println(ui.StyleSuccess.Render("✓ finalized"))

	// Update idea status
	fmt.Print(ui.StyleMuted.Render("Updating idea status... "))
	if err := repo.UpdateIdeaStatus(script.IdeaID, "finalized"); err != nil {
		fmt.Println(ui.StyleWarning.Render("⚠ " + err.Error()))
	} else {
		fmt.Println(ui.StyleSuccess.Render("✓ finalized"))
	}

	fmt.Printf("\n%s\n", ui.StyleSuccess.Render("✅ Script finalized!"))
	fmt.Println("\nNext steps:")
	fmt.Printf("  • Export file: %s\n", exportPath)
	fmt.Println("  • Create calendar plan: gagipress calendar plan")
	fmt.Println("  • Or export from calendar: gagipress calendar export <id>")

	return nil
}

func saveExportFile(script *models.ContentScript, platform, caption string) (string, error) {
	// Ensure exports/ directory exists
	if err := os.MkdirAll("exports", 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	// Build filename
	date := time.Now().Format("2006-01-02")
	shortID := script.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	filename := fmt.Sprintf("%s-%s-%s.md", date, platform, shortID)
	exportPath := filepath.Join("exports", filename)

	// Build markdown content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s Post — %s\n\n", strings.Title(platform), date))
	content.WriteString("## Caption\n\n")
	content.WriteString(caption)
	content.WriteString("\n\n## Production Notes\n\n")
	content.WriteString(fmt.Sprintf("**Duration:** %ds\n\n", script.EstimatedDuration))
	content.WriteString("---\nGenerated by Gagipress CLI\n")

	if err := os.WriteFile(exportPath, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return exportPath, nil
}
```

**Important:** The `saveExportFile` function uses `models.ContentScript` — add this import:
```go
"github.com/gagipress/gagipress-cli/internal/models"
```

**Note on `strings.Title`:** This is deprecated in Go 1.18+. Use `cases.Title(language.English).String(platform)` from `golang.org/x/text` instead, or simply capitalize manually:

```go
// Replace strings.Title(platform) with:
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
```

Add `capitalize` as a package-level function in `finalize.go` and use `capitalize(platform)` instead of `strings.Title(platform)`.

**Step 2: Verify it compiles**

Run: `make build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add cmd/scripts/finalize.go
git commit -m "feat(scripts): add finalize command with export"
```

---

### Task 8: Update scheduler to use only finalized scripts

**Files:**
- Modify: `internal/scheduler/planner.go:29`

**Step 1: Change GetScripts call to filter by finalized**

In `internal/scheduler/planner.go`, line 29, change:

```go
scripts, err := p.contentRepo.GetScripts("", 0)
```

to:

```go
scripts, err := p.contentRepo.GetScripts("finalized", 0)
```

Also update the error message on line 36 from:
```go
return nil, fmt.Errorf("no scripts available for planning")
```
to:
```go
return nil, fmt.Errorf("no finalized scripts available for planning (finalize scripts first with 'gagipress scripts finalize')")
```

**Step 2: Verify it compiles**

Run: `make build`
Expected: Build succeeds

**Step 3: Run tests**

Run: `make test`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/scheduler/planner.go
git commit -m "feat(scheduler): only use finalized scripts for calendar planning"
```

---

### Task 9: Create `calendar export` command

**Files:**
- Create: `cmd/calendar/export.go`

**Step 1: Create the export command**

Create `cmd/calendar/export.go`:

```go
package calendar

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var exportWeek bool

var exportCmd = &cobra.Command{
	Use:   "export [calendar-id]",
	Short: "Export calendar entries as publish-ready files",
	Long: `Export approved calendar entries as markdown files ready for publishing.

Use with a specific ID to export one entry, or --week to export all approved entries.`,
	RunE: runExport,
}

func init() {
	exportCmd.Flags().BoolVar(&exportWeek, "week", false, "Export all approved entries for the current week")

	CalendarCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	if !exportWeek && len(args) == 0 {
		return fmt.Errorf("provide a calendar entry ID or use --week flag")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("📤 Export Calendar Entries"))
	fmt.Println()

	contentRepo := repository.NewContentRepository(&cfg.Supabase)
	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)

	if exportWeek {
		return exportWeekEntries(calendarRepo, contentRepo)
	}

	return exportSingleEntry(args[0], calendarRepo, contentRepo)
}

func exportSingleEntry(entryID string, calendarRepo *repository.CalendarRepository, contentRepo *repository.ContentRepository) error {
	// Get all approved entries and find the matching one
	entries, err := calendarRepo.GetEntries("approved", 0)
	if err != nil {
		return fmt.Errorf("failed to get calendar entries: %w", err)
	}

	var entry *models.ContentCalendar
	for i := range entries {
		if entries[i].ID == entryID || (len(entries[i].ID) >= len(entryID) && entries[i].ID[:len(entryID)] == entryID) {
			entry = &entries[i]
			break
		}
	}

	if entry == nil {
		return fmt.Errorf("approved calendar entry not found: %s", entryID)
	}

	path, err := exportCalendarEntry(entry, contentRepo)
	if err != nil {
		return err
	}

	fmt.Printf("%s Exported: %s\n", ui.StyleSuccess.Render("✓"), path)
	return nil
}

func exportWeekEntries(calendarRepo *repository.CalendarRepository, contentRepo *repository.ContentRepository) error {
	entries, err := calendarRepo.GetEntries("approved", 0)
	if err != nil {
		return fmt.Errorf("failed to get calendar entries: %w", err)
	}

	// Filter to current week
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	var weekEntries []models.ContentCalendar
	for _, entry := range entries {
		if entry.ScheduledFor.After(weekStart) && entry.ScheduledFor.Before(weekEnd) {
			weekEntries = append(weekEntries, entry)
		}
	}

	if len(weekEntries) == 0 {
		fmt.Println("No approved calendar entries found for this week.")
		fmt.Println("\nSchedule content with: gagipress calendar plan")
		return nil
	}

	exported := 0
	for i := range weekEntries {
		path, err := exportCalendarEntry(&weekEntries[i], contentRepo)
		if err != nil {
			fmt.Printf("%s Failed to export %s: %v\n", ui.StyleError.Render("✗"), weekEntries[i].ID[:8], err)
			continue
		}
		fmt.Printf("%s Exported: %s\n", ui.StyleSuccess.Render("✓"), path)
		exported++
	}

	fmt.Printf("\n%s %d/%d entries exported\n", ui.StyleSuccess.Render("✅"), exported, len(weekEntries))
	return nil
}

func exportCalendarEntry(entry *models.ContentCalendar, contentRepo *repository.ContentRepository) (string, error) {
	if entry.ScriptID == nil {
		return "", fmt.Errorf("calendar entry %s has no associated script", entry.ID[:8])
	}

	script, err := contentRepo.GetScriptByID(*entry.ScriptID)
	if err != nil {
		return "", fmt.Errorf("failed to get script: %w", err)
	}

	// Compose caption
	hashtagStr := strings.Join(script.Hashtags, " ")
	caption := script.Hook + "\n\n" + script.CTA + "\n\n" + hashtagStr

	// Ensure exports/ directory exists
	if err := os.MkdirAll("exports", 0755); err != nil {
		return "", fmt.Errorf("failed to create exports directory: %w", err)
	}

	// Build filename
	date := entry.ScheduledFor.Format("2006-01-02")
	shortID := entry.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	filename := fmt.Sprintf("%s-%s-%s.md", date, entry.Platform, shortID)
	exportPath := filepath.Join("exports", filename)

	// Build markdown
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s Post — %s\n\n", capitalize(entry.Platform), date))
	content.WriteString("## Caption\n\n")
	content.WriteString(caption)
	content.WriteString("\n\n## Production Notes\n\n")
	content.WriteString(fmt.Sprintf("**Duration:** %ds\n", script.EstimatedDuration))
	content.WriteString(fmt.Sprintf("**Scheduled:** %s\n", entry.ScheduledFor.Format("Mon Jan 02, 15:04")))
	content.WriteString(fmt.Sprintf("**Post Type:** %s\n\n", entry.PostType))
	content.WriteString("---\nGenerated by Gagipress CLI\n")

	if err := os.WriteFile(exportPath, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return exportPath, nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
```

**Step 2: Verify it compiles**

Run: `make build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add cmd/calendar/export.go
git commit -m "feat(calendar): add export command for publish-ready files"
```

---

### Task 10: Add `exports/` to .gitignore and final build check

**Files:**
- Modify: `.gitignore`

**Step 1: Add exports/ to .gitignore**

Add after the `dist/` line in `.gitignore`:

```
exports/
```

**Step 2: Full build and test**

Run: `make build && make test`
Expected: Both succeed

**Step 3: Commit**

```bash
git add .gitignore
git commit -m "chore: add exports/ to gitignore"
```

---

## Summary of all tasks

| # | Task | Files | Type |
|---|------|-------|------|
| 1 | Add Status to ContentScript | models/content.go | Modify |
| 2 | Add FormatStatus badges | ui/styles.go | Modify |
| 3 | Add repo methods (UpdateScriptStatus, GetScriptByID, status filter) | repository/content.go, content_test.go, planner.go | Modify + Test |
| 4 | SQL migration find_script_by_prefix | migrations/004*, supabase/migrations/004* | Create |
| 5 | Add GetScriptByIDPrefix repo method | repository/content.go, content_test.go | Modify + Test |
| 6 | Create scripts command group + list | cmd/scripts/*.go, cmd/root.go | Create + Modify |
| 7 | Create scripts finalize command | cmd/scripts/finalize.go | Create |
| 8 | Update scheduler to use finalized | scheduler/planner.go | Modify |
| 9 | Create calendar export command | cmd/calendar/export.go | Create |
| 10 | Add exports/ to gitignore | .gitignore | Modify |
