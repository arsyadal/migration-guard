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
