package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
	FilePath    string
}

// LoadMigrations loads all migration files from the migrations directory
func LoadMigrations(migrationsDir string) ([]Migration, error) {
	var migrations []Migration

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		filePath := filepath.Join(migrationsDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		// Extract version from filename (e.g., "001_initial_schema.sql" -> 1)
		var version int
		var description string
		parts := strings.SplitN(file.Name(), "_", 2)
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &version)
			description = strings.TrimSuffix(parts[1], ".sql")
			description = strings.ReplaceAll(description, "_", " ")
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
			FilePath:    filePath,
		})
	}

	return migrations, nil
}

// RunMigration executes a migration SQL using Supabase REST API
func (c *Client) RunMigration(migration Migration) error {
	// Supabase doesn't have a direct "execute SQL" endpoint in the REST API
	// We need to use the PostgREST admin API or pg_net extension
	// For now, we'll use a workaround: execute via HTTP POST to /rest/v1/rpc

	// Split SQL into individual statements
	statements := splitSQLStatements(migration.SQL)

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// Execute each statement via raw SQL
		if err := c.executeRawSQL(stmt); err != nil {
			return fmt.Errorf("failed to execute statement %d: %w\nStatement: %s", i+1, err, stmt)
		}
	}

	return nil
}

// executeRawSQL executes raw SQL using Supabase's PostgREST
func (c *Client) executeRawSQL(sql string) error {
	// Build request URL
	url := fmt.Sprintf("%s/rest/v1/rpc/exec_sql", c.config.URL)

	// Prepare request body
	body := map[string]interface{}{
		"sql": sql,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	apiKey := c.config.ServiceKey
	if apiKey == "" {
		apiKey = c.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SQL execution failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// splitSQLStatements splits a SQL file into individual statements
func splitSQLStatements(sql string) []string {
	// Simple split by semicolon (not perfect but works for most cases)
	// This doesn't handle complex cases like semicolons in strings or function bodies
	statements := strings.Split(sql, ";")

	var result []string
	var buffer strings.Builder
	inFunctionOrBlock := false

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)

		// Check if we're entering a function/procedure/block
		lowerStmt := strings.ToLower(stmt)
		if strings.Contains(lowerStmt, "create function") ||
			strings.Contains(lowerStmt, "create or replace function") ||
			strings.Contains(lowerStmt, "create procedure") ||
			strings.Contains(lowerStmt, "do $$") {
			inFunctionOrBlock = true
		}

		// Add to buffer
		if buffer.Len() > 0 {
			buffer.WriteString("; ")
		}
		buffer.WriteString(stmt)

		// Check if we're closing a function/procedure/block
		if inFunctionOrBlock {
			if strings.Contains(lowerStmt, "language plpgsql") ||
				strings.Contains(lowerStmt, "$$") {
				inFunctionOrBlock = false
				result = append(result, buffer.String())
				buffer.Reset()
			}
		} else {
			// Regular statement
			if stmt != "" {
				result = append(result, buffer.String())
				buffer.Reset()
			}
		}
	}

	// Add any remaining buffer content
	if buffer.Len() > 0 {
		result = append(result, buffer.String())
	}

	return result
}

// GetAppliedVersion checks which migrations have been applied
func (c *Client) GetAppliedVersion() (int, error) {
	// Try to query the schema_version table
	// If it doesn't exist, we'll get an error and return 0

	url := fmt.Sprintf("%s/rest/v1/schema_version?select=version&order=version.desc&limit=1", c.config.URL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := c.config.ServiceKey
	if apiKey == "" {
		apiKey = c.config.AnonKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil // Table might not exist
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode >= 400 {
		return 0, nil // Table doesn't exist yet
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result) == 0 {
		return 0, nil
	}

	version, ok := result[0]["version"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid version type")
	}

	return int(version), nil
}
