package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHarness(t *testing.T) {
	if false {
		t.Error("harness failed")
	}
}

func TestAnalyzeSQL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	upFile := filepath.Join(tmpDir, "0001_test.up.sql")
	sqlContent := `
		-- This is a comment
		/* This is a multi-line
		   comment */
		ALTER TABLE users RENAME COLUMN old_name TO new_name;
		ALTER TABLE users ADD COLUMN new_col VARCHAR(255) NOT NULL;
		CREATE TABLE logs (id INT);
	`
	if err := os.WriteFile(upFile, []byte(sqlContent), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := AnalyzeSQL(upFile)
	if err != nil {
		t.Fatalf("AnalyzeSQL failed: %v", err)
	}

	var hasRename, hasNullNoDefault, hasNoPK bool
	for _, res := range results {
		if res.Status == StatusCritical && res.CheckItem == "Rename Column" {
			hasRename = true
		}
		if res.Status == StatusWarning && res.CheckItem == "Column Constraints" {
			hasNullNoDefault = true
		}
		if res.Status == StatusWarning && res.CheckItem == "Primary Key Check" {
			hasNoPK = true
		}
	}

	if !hasRename {
		t.Error("Expected Critical check for RENAME COLUMN")
	}
	if !hasNullNoDefault {
		t.Error("Expected Warning check for ADD COLUMN NOT NULL without DEFAULT")
	}
	if !hasNoPK {
		t.Error("Expected Warning check for CREATE TABLE without PRIMARY KEY")
	}
}

func TestScanDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	files := []string{
		"0001_init.up.sql",
		"0001_init.down.sql",
		"0002_add_users.up.sql",
		"0003_add_logs.down.sql", // Missing up
		"ignored_file.txt",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("SELECT 1;"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	pairs, err := ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(pairs) != 3 {
		t.Fatalf("Expected 3 migration pairs/files, got %d", len(pairs))
	}

	// Verify order is sorted
	if pairs[0].BaseName != "0001_init" {
		t.Errorf("Expected 0001_init, got %s", pairs[0].BaseName)
	}
}

func TestIntegrationCLI(t *testing.T) {
	// Simple validation of down/rollback verification
	tmpDir, err := os.MkdirTemp("", "migrations")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create up migration but no down migration (critical fail)
	if err := os.WriteFile(filepath.Join(tmpDir, "0001_init.up.sql"), []byte("CREATE TABLE users (id INT PRIMARY KEY);"), 0644); err != nil {
		t.Fatal(err)
	}

	pairs, err := ScanDirectory(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	var results []AuditResult
	for _, pair := range pairs {
		if pair.UpPath != "" {
			res, err := AnalyzeSQL(pair.UpPath)
			if err != nil {
				t.Fatal(err)
			}
			results = append(results, res...)
		}
		// Verify rollback presence
		if pair.DownPath == "" {
			results = append(results, AuditResult{
				File:      pair.BaseName + ".down.sql",
				CheckItem: "Rollback Verification",
				Status:    StatusCritical,
				Details:   "Corresponding down migration file is missing or empty.",
			})
		}
	}

	var criticalCount int
	for _, r := range results {
		if r.Status == StatusCritical {
			criticalCount++
		}
	}

	if criticalCount != 1 {
		t.Errorf("Expected 1 critical violation (missing rollback), got %d", criticalCount)
	}
}


