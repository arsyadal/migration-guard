# Design Doc: Database Schema Safety Guardrail for AI Agents

## 1. Objective
A lightweight, zero-dependency CLI tool written in Go that acts as a pre-commit/pre-triage guardrail for SQL migrations. It scans SQL files, identifies destructive or inefficient operations, verifies rollback scripts, and outputs a structured Markdown compliance report.

## 2. Component Design

### Migration Scanner (`migration_scanner.go`)
- Function `ScanDirectory(dir string) ([]MigrationPair, error)`
- Scans `dir` for `.sql` files.
- Groups migrations by base name or version prefix.
- Returns a list of pairs containing the paths to `.up.sql` and `.down.sql` files.

### SQL Analyzer (`sql_analyzer.go`)
- Function `AnalyzeSQL(filePath string) ([]AuditResult, error)`
- Strips single-line (`--`) and multi-line (`/* ... */`) SQL comments.
- Runs regex analysis to detect:
  - **Critical Hazards**:
    - `DROP TABLE`
    - `DROP COLUMN`
    - `ALTER TABLE ... RENAME COLUMN`
  - **Warnings**:
    - `ADD COLUMN ... NOT NULL` without `DEFAULT`
    - `CREATE TABLE` without `PRIMARY KEY`
- Rollback verification: Checks if each `.up.sql` has a corresponding `.down.sql` file that is not empty.

### CLI Entrypoint (`main.go`)
- Parses `-dir` flag (defaults to `./migrations`).
- Triggers scan and analysis.
- Renders Markdown report to stdout.
- Exits with code `1` if any **Critical** checks fail, otherwise exits with code `0`.

## 3. Output Format
Renders a clean Markdown table with:
- Migration File
- Check Item
- Status (Green / ⚠️ Warning / ❌ Critical)
- Details / Remediation
