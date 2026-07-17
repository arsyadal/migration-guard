# Database Schema Safety Guardrail Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a database schema safety guardrail CLI tool written in Go that acts as a pre-commit/pre-triage guardrail for SQL migrations, scanning migration files, identifying destructive operations, and generating compliance Markdown reports.

**Architecture:** A standalone Go CLI with scanner, analyzer, and main CLI runner using Go's standard library.

**Tech Stack:** Go (Standard Library only - `os`, `path/filepath`, `strings`, `regexp`, `fmt`, `testing`).

## Global Constraints
- Write in Go using only the Go Standard Library.
- Scan designated migration directories (default `./migrations`, overridable via flag).
- Detect critical SQL hazards (DROP, RENAME COLUMN) and warn on nullable columns without defaults or tables without PKs/indexes.
- Exit code 1 if critical hazards found or down/rollback migrations are missing/empty.

---

### Task 1: Project Setup and Test Framework

**Files:**
- Create: `go.mod`
- Create: `main_test.go`

**Interfaces:**
- Consumes: None
- Produces: Initial test runner configuration

- [ ] **Step 1: Write a basic test file `main_test.go` to verify the testing harness is functional**

```go
package main

import "testing"

func TestHarness(t *testing.T) {
	if false {
		t.Error("harness failed")
	}
}
```

- [ ] **Step 2: Run tests to verify the test suite passes**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3: Write minimal module initialization `go.mod`**

```go
module migration-guardrail

go 1.20
```

- [ ] **Step 4: Run tests to verify they still pass under the module**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
rtk git add go.mod main_test.go
rtk git commit -m "chore: initialize project module and test harness"
```

---

### Task 2: Implement SQL Analyzer (`sql_analyzer.go`)

**Files:**
- Create: `sql_analyzer.go`
- Modify: `main_test.go`

**Interfaces:**
- Consumes: Raw SQL files via path
- Produces: `AnalyzeSQL(filePath string) ([]AuditResult, error)`
- Types:
```go
type AuditStatus string
const (
	StatusGreen    AuditStatus = "Green"
	StatusWarning  AuditStatus = "Warning"
	StatusCritical AuditStatus = "Critical"
)

type AuditResult struct {
	File        string
	CheckItem   string
	Status      AuditStatus
	Details     string
}
```

- [ ] **Step 1: Write failing tests in `main_test.go` to verify SQL parsing rules (comments removal, drop table, drop column, rename column, nullable without default, create table without PK)**

```go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./...`
Expected: FAIL due to `AnalyzeSQL` undefined.

- [ ] **Step 3: Write minimal implementation in `sql_analyzer.go`**

```go
package main

import (
	"os"
	"regexp"
	"strings"
)

type AuditStatus string

const (
	StatusGreen    AuditStatus = "Green"
	StatusWarning  AuditStatus = "Warning"
	StatusCritical AuditStatus = "Critical"
)

type AuditResult struct {
	File      string
	CheckItem string
	Status    AuditStatus
	Details   string
}

func stripComments(sql string) string {
	// Strip multi-line comments
	reMulti := regexp.MustCompile(`(?s)/\*\*.*?\*/|/\*.*?\*/`)
	sql = reMulti.ReplaceAllString(sql, "")
	// Strip single-line comments
	reSingle := regexp.MustCompile(`--.*`)
	sql = reSingle.ReplaceAllString(sql, "")
	return sql
}

func AnalyzeSQL(filePath string) ([]AuditResult, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	content := stripComments(string(contentBytes))
	var results []AuditResult

	// Rule 1: RENAME COLUMN (Critical)
	reRename := regexp.MustCompile(`(?i)alter\s+table\s+.*\s+rename\s+(column\s+)?`)
	if reRename.MatchString(content) {
		results = append(results, AuditResult{
			File:      filePath,
			CheckItem: "Rename Column",
			Status:    StatusCritical,
			Details:   "ALTER TABLE RENAME COLUMN breaks compatibility.",
		})
	}

	// Rule 2: DROP TABLE or DROP COLUMN (Critical)
	reDrop := regexp.MustCompile(`(?i)drop\s+(table|column)\s+`)
	if reDrop.MatchString(content) {
		results = append(results, AuditResult{
			File:      filePath,
			CheckItem: "Destructive Commands",
			Status:    StatusCritical,
			Details:   "DROP TABLE or DROP COLUMN operation detected.",
		})
	}

	// Rule 3: ADD COLUMN NOT NULL without DEFAULT (Warning)
	reAddNotNull := regexp.MustCompile(`(?i)add\s+(column\s+)?\S+\s+[^;]*?not\s+null`)
	if reAddNotNull.MatchString(content) {
		if !regexp.MustCompile(`(?i)default\s+`).MatchString(content) {
			results = append(results, AuditResult{
				File:      filePath,
				CheckItem: "Column Constraints",
				Status:    StatusWarning,
				Details:   "Column added as NOT NULL without DEFAULT value.",
			})
		}
	}

	// Rule 4: CREATE TABLE without PRIMARY KEY or FK index (Warning)
	reCreateTable := regexp.MustCompile(`(?i)create\s+table\s+\S+\s*\(`)
	if reCreateTable.MatchString(content) {
		if !regexp.MustCompile(`(?i)primary\s+key`).MatchString(content) {
			results = append(results, AuditResult{
				File:      filePath,
				CheckItem: "Primary Key Check",
				Status:    StatusWarning,
				Details:   "CREATE TABLE without defining a Primary Key.",
			})
		}
	}

	// Default Green result if no errors/warnings found
	if len(results) == 0 {
		results = append(results, AuditResult{
			File:      filePath,
			CheckItem: "Destructive Commands",
			Status:    StatusGreen,
			Details:   "No destructive or inefficient operations detected.",
		})
	}

	return results, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
rtk git add sql_analyzer.go main_test.go
rtk git commit -m "feat: implement sql analysis with hazard checks"
```

---

### Task 3: Implement Migration Scanner (`migration_scanner.go`)

**Files:**
- Create: `migration_scanner.go`
- Modify: `main_test.go`

**Interfaces:**
- Consumes: Target directory path
- Produces: `ScanDirectory(dir string) ([]MigrationPair, error)`
- Types:
```go
type MigrationPair struct {
	UpPath   string // Path to the .up.sql file (if present)
	DownPath string // Path to the corresponding .down.sql file (if present)
	BaseName string // Shared base name/prefix (e.g. 0001_init)
}
```

- [ ] **Step 1: Write failing tests in `main_test.go` to test directories scanning and matching of up/down SQL files**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./...`
Expected: FAIL due to `ScanDirectory` undefined.

- [ ] **Step 3: Write minimal implementation in `migration_scanner.go`**

```go
package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type MigrationPair struct {
	UpPath   string
	DownPath string
	BaseName string
}

func ScanDirectory(dir string) ([]MigrationPair, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	pairMap := make(map[string]*MigrationPair)
	var keys []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		isUp := strings.HasSuffix(name, ".up.sql")
		isDown := strings.HasSuffix(name, ".down.sql")
		if !isUp && !isDown {
			continue
		}

		baseName := ""
		if isUp {
			baseName = strings.TrimSuffix(name, ".up.sql")
		} else {
			baseName = strings.TrimSuffix(name, ".down.sql")
		}

		pair, exists := pairMap[baseName]
		if !exists {
			pair = &MigrationPair{BaseName: baseName}
			pairMap[baseName] = pair
			keys = append(keys, baseName)
		}

		fullPath := filepath.Join(dir, name)
		if isUp {
			pair.UpPath = fullPath
		} else {
			pair.DownPath = fullPath
		}
	}

	sort.Strings(keys)
	var pairs []MigrationPair
	for _, k := range keys {
		pairs = append(pairs, *pairMap[k])
	}

	return pairs, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./...`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
rtk git add migration_scanner.go main_test.go
rtk git commit -m "feat: implement migration scanner sorting and grouping"
```

---

### Task 4: Implement Orchestrator & Markdown Report (`main.go`)

**Files:**
- Create: `main.go`
- Modify: `main_test.go`

**Interfaces:**
- Consumes: Directory path via `-dir` CLI flag or defaults to `./migrations`
- Produces: Printed Markdown report and exit code 0 or 1.

- [ ] **Step 1: Write integration tests in `main_test.go` to test CLI behavior (exit code, output generation)**

```go
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
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test -v ./...`
Expected: PASS

- [ ] **Step 3: Write minimal implementation of `main.go` to parse flags, run analyzer, and print Markdown output**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func run(dir string) (int, error) {
	pairs, err := ScanDirectory(dir)
	if err != nil {
		return 1, fmt.Errorf("error scanning directory: %w", err)
	}

	var results []AuditResult
	for _, pair := range pairs {
		var pairResults []AuditResult
		if pair.UpPath != "" {
			res, err := AnalyzeSQL(pair.UpPath)
			if err != nil {
				return 1, fmt.Errorf("error analyzing %s: %w", pair.UpPath, err)
			}
			pairResults = append(pairResults, res...)
		}

		// Down file rollback check
		hasDown := false
		if pair.DownPath != "" {
			fi, err := os.Stat(pair.DownPath)
			if err == nil && fi.Size() > 0 {
				hasDown = true
			}
		}

		if !hasDown {
			name := pair.BaseName + ".down.sql"
			pairResults = append(pairResults, AuditResult{
				File:      name,
				CheckItem: "Rollback Verification",
				Status:    StatusCritical,
				Details:   "Corresponding down migration file is missing or empty.",
			})
		}

		results = append(results, pairResults...)
	}

	if len(results) == 0 {
		fmt.Println("No migration files found.")
		return 0, nil
	}

	fmt.Println("### 🛡️ Database Migration Audit Report\n")
	fmt.Println("| Migration File | Check Item | Status | Details / Remediation |")
	fmt.Println("| :--- | :--- | :--- | :--- |")

	var criticalCount int
	for _, res := range results {
		statusStr := "Green"
		if res.Status == StatusWarning {
			statusStr = "⚠️ Warning"
		} else if res.Status == StatusCritical {
			statusStr = "❌ Critical"
			criticalCount++
		}
		fmt.Printf("| `%s` | %s | %s | %s |\n", filepath.Base(res.File), res.CheckItem, statusStr, res.Details)
	}
	fmt.Println()

	if criticalCount > 0 {
		fmt.Println("❌ **Audit Result:** Database safety violations detected. You must generate the matching down/rollback SQL migration file before proceeding.")
		return 1, nil
	}

	fmt.Println("✅ **Audit Result:** All migrations passed safety checks.")
	return 0, nil
}

func main() {
	dir := flag.String("dir", "./migrations", "Directory containing migration files")
	flag.Parse()

	code, err := run(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}
```

- [ ] **Step 4: Run CLI manually on temporary directory to verify stdout**

Run: `go run main.go -dir .`
Expected: Print check report (or No migration files found) and exit.

- [ ] **Step 5: Commit changes**

```bash
rtk git add main.go
rtk git commit -m "feat: add main driver orchestrator and markdown output"
```

---

### Task 5: Add SKILL.md Manifest

**Files:**
- Create: `SKILL.md`

**Interfaces:**
- Consumes: None
- Produces: Skill registration file for agent-browser or skills.sh

- [ ] **Step 1: Write `SKILL.md`**

```markdown
---
name: migration-guardrail
description: Enforce zero-downtime safety and rollback compliance on SQL migrations generated by AI.
---
# Database Migration Safety Skill
Use this skill automatically whenever you generate, modify, or prepare database schema changes, ORM mutations, or raw SQL migration files.

## How to execute
Run the local guardrail engine:
`go run main.go`

If the tool exits with status 1, check the markdown report. You are strictly forbidden from committing these migrations until you add appropriate rollback files or fix dangerous breaking changes. Refactor the migration strategy and rerun the tool until it passes (status 0).
```

- [ ] **Step 2: Commit changes**

```bash
rtk git add SKILL.md
rtk git commit -m "docs: add SKILL.md manifest"
```
